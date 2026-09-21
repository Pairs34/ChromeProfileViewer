package launcher

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pairs/chrome-profile-viewer/internal/model"
)

type Service struct {
	browsers map[string]model.Browser
	tempRoot string
}

func New(browsers []model.Browser) *Service {
	lookup := make(map[string]model.Browser, len(browsers))
	for _, browser := range browsers {
		lookup[browser.ID] = browser
	}
	service := &Service{browsers: lookup, tempRoot: filepath.Join(os.TempDir(), "browser-profile-viewer")}
	service.removeExpiredCopies(7 * 24 * time.Hour)
	return service
}

func (s *Service) removeExpiredCopies(maxAge time.Duration) {
	entries, err := os.ReadDir(s.tempRoot)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "session-") {
			continue
		}
		info, err := entry.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			_ = os.RemoveAll(filepath.Join(s.tempRoot, entry.Name()))
		}
	}
}

func (s *Service) Launch(profile model.Profile, browserID string, isolated bool) (model.LaunchResult, error) {
	browser, ok := s.browsers[browserID]
	if !ok {
		return model.LaunchResult{}, errors.New("seçilen tarayıcı bulunamadı")
	}
	if browser.Family != profile.Family {
		return model.LaunchResult{}, fmt.Errorf("%s profili %s ile açılamaz", profile.Family, browser.Name)
	}
	if _, err := os.Stat(profile.Path); err != nil {
		return model.LaunchResult{}, fmt.Errorf("profil klasörü okunamıyor: %w", err)
	}

	effective := profile
	if isolated {
		copyProfile, err := s.isolatedCopy(profile)
		if err != nil {
			return model.LaunchResult{}, fmt.Errorf("güvenli kopya oluşturulamadı: %w", err)
		}
		effective = copyProfile
	}
	args, err := BuildArguments(effective)
	if err != nil {
		return model.LaunchResult{}, err
	}
	if effective.Family == model.FamilyFirefox {
		// Eski Firefox "profil daha yeni sürümle kullanılmış" uyarısı vermesin diye
		// sürüm bilgisini tutan dosyayı sileriz; Firefox açılışta yeniden oluşturur.
		if err := os.Remove(filepath.Join(effective.Path, "compatibility.ini")); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return model.LaunchResult{}, fmt.Errorf("compatibility.ini silinemedi (Firefox açık olabilir): %w", err)
		}
		if isolated {
			args = append(args, "-allow-downgrade")
		}
	}
	command := exec.Command(browser.Executable, args...)
	if err := command.Start(); err != nil {
		return model.LaunchResult{}, fmt.Errorf("tarayıcı başlatılamadı: %w", err)
	}
	return model.LaunchResult{
		PID: command.Process.Pid, Browser: browser.Name, Profile: profile.Name,
		EffectiveDir: effective.UserDataDir, Isolated: isolated, Arguments: args,
	}, nil
}

func BuildArguments(profile model.Profile) ([]string, error) {
	switch profile.Family {
	case model.FamilyChromium:
		if profile.UserDataDir == "" || profile.DirectoryName == "" {
			return nil, errors.New("Chromium profil bilgisi eksik")
		}
		return []string{"--user-data-dir=" + profile.UserDataDir, "--profile-directory=" + profile.DirectoryName, "--new-window"}, nil
	case model.FamilyFirefox:
		if profile.Path == "" {
			return nil, errors.New("Firefox profil yolu eksik")
		}
		return []string{"-no-remote", "-profile", profile.Path}, nil
	default:
		return nil, errors.New("desteklenmeyen profil türü")
	}
}

func (s *Service) isolatedCopy(profile model.Profile) (model.Profile, error) {
	if err := os.MkdirAll(s.tempRoot, 0o700); err != nil {
		return profile, err
	}
	session, err := os.MkdirTemp(s.tempRoot, "session-")
	if err != nil {
		return profile, err
	}
	if profile.Family == model.FamilyChromium {
		name := profile.DirectoryName
		if name == "" {
			name = "Default"
		}
		destination := filepath.Join(session, name)
		if err := copyDirectory(profile.Path, destination); err != nil {
			return profile, err
		}
		localState := filepath.Join(profile.UserDataDir, "Local State")
		if _, err := os.Stat(localState); err == nil {
			if err := copyFile(localState, filepath.Join(session, "Local State"), 0o600); err != nil {
				return profile, err
			}
		}
		profile.Path, profile.UserDataDir, profile.DirectoryName = destination, session, name
		profile.Locked = false
		return profile, nil
	}
	destination := filepath.Join(session, "profile")
	if err := copyDirectory(profile.Path, destination); err != nil {
		return profile, err
	}
	profile.Path, profile.UserDataDir, profile.Locked = destination, destination, false
	return profile, nil
}

var skippedCopyNames = map[string]bool{
	"cache": true, "code cache": true, "gpucache": true, "shadercache": true,
	"dawncache": true, "grshadercache": true, "crashpad": true,
	"singletonlock": true, "singletoncookie": true, "singletonsocket": true,
	"parent.lock": true, ".parentlock": true, "lock": true,
	// Firefox: son çalışan sürümü tutar; eski Firefox "yeni profil oluştur" der.
	"compatibility.ini": true, "startupcache": true, "cache2": true,
}

func copyDirectory(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return os.MkdirAll(destination, 0o700)
		}
		if skippedCopyNames[strings.ToLower(entry.Name())] {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(source, destination string, mode fs.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
