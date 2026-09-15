# Browser Profile Viewer

Chrome/Chromium ve Firefox tabanlı yerel profilleri bulan ve uygun yerel tarayıcıyla açan Go/Wails masaüstü uygulaması.

Desteklenen ürünler: Google Chrome, Microsoft Edge, Brave, Chromium, Vivaldi,
Opera, Mozilla Firefox ve LibreWolf. Safari özel profil klasörüyle başlatmayı
desteklemediği için kapsam dışındadır.

## Geliştirme

Gereksinimler: Go 1.25+ ve Wails v2.15 CLI.

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.15.0
wails dev
```

Test ve üretim derlemesi:

```bash
go test ./...
wails build
```

Güncel macOS SDK'sında doğrudan uygulama paketi üretmek için:

```bash
./scripts/build-macos.sh
open build/bin/BrowserProfileViewer.app
```

macOS paketi, bu projede kullanılan Go 1.25+ araç zinciri nedeniyle macOS 13+
için üretilir. Windows üretim paketi `wails build` ile oluşturulabilir.

Eski .NET Framework/WinForms uygulaması `old/` klasöründe korunmaktadır.
