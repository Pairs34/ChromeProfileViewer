const state = {
  browsers: [],
  profiles: [],
  selected: new Set(),
  browserChoices: new Map(),
  filter: "all",
  query: "",
};

const byId = (id) => document.getElementById(id);

function backend() {
  if (!window.go?.main?.App) throw new Error("Uygulama bağlantısı hazır değil.");
  return window.go.main.App;
}

function toast(message, error = false) {
  const element = byId("toast");
  element.textContent = message;
  element.classList.toggle("error", error);
  element.hidden = false;
  clearTimeout(toast.timer);
  toast.timer = setTimeout(() => { element.hidden = true; }, 4500);
}

function setScanning(busy) {
  byId("chooseButton").disabled = busy;
  byId("scanButton").disabled = busy;
  byId("scanButton").textContent = busy ? "Taranıyor" : "Tara";
}

async function refreshBrowsers() {
  try {
    state.browsers = await backend().DetectBrowsers();
    const names = state.browsers.map((item) => item.name);
    byId("browserStatus").textContent = names.length ? names.join(" · ") : "Tarayıcı bulunamadı";
    byId("browserStatus").title = names.join(", ");
  } catch (error) {
    byId("browserStatus").textContent = "Tarayıcı algılanamadı";
    toast(String(error), true);
  }
}

async function chooseDirectory() {
  try {
    const path = await backend().SelectDirectory();
    if (!path) return;
    byId("pathInput").value = path;
    await scan(path);
  } catch (error) {
    toast(String(error), true);
  }
}

async function scan(path = byId("pathInput").value.trim()) {
  if (!path) {
    toast("Önce bir klasör seçin.", true);
    return;
  }
  setScanning(true);
  try {
    const result = await backend().ScanDirectory(path);
    state.browsers = result.browsers || [];
    state.profiles = result.profiles || [];
    state.selected.clear();
    state.browserChoices.clear();
    for (const profile of state.profiles) {
      state.browserChoices.set(profile.id, profile.suggestedBrowserId || firstCompatibleBrowser(profile)?.id || "");
    }
    byId("pathInput").value = result.root;
    showWarnings(result.warnings || []);
    render();
  } catch (error) {
    toast(String(error), true);
  } finally {
    setScanning(false);
  }
}

function showWarnings(items) {
  const element = byId("warnings");
  element.hidden = items.length === 0;
  element.textContent = items.join(" • ");
}

function visibleProfiles() {
  const query = state.query.toLocaleLowerCase("tr");
  return state.profiles.filter((profile) => {
    const familyMatches = state.filter === "all" || profile.family === state.filter;
    const text = `${profile.name} ${profile.path} ${profile.productHint}`.toLocaleLowerCase("tr");
    return familyMatches && (!query || text.includes(query));
  });
}

function render() {
  const visible = visibleProfiles();
  const container = byId("profiles");
  const table = byId("profileTable");
  const empty = byId("emptyState");
  container.replaceChildren();

  byId("summary").textContent = state.profiles.length
    ? `${state.profiles.length} profil bulundu${visible.length !== state.profiles.length ? ` · ${visible.length} gösteriliyor` : ""}`
    : "Desteklenen profil bulunamadı.";

  table.hidden = visible.length === 0;
  empty.hidden = visible.length !== 0;
  if (!state.profiles.length) {
    empty.querySelector("p").textContent = "Bu klasörde profil bulunamadı.";
  } else if (!visible.length) {
    empty.querySelector("p").textContent = "Filtreyle eşleşen profil yok.";
  }

  for (const profile of visible) container.appendChild(profileRow(profile));
  updateSelection(visible);
}

function profileRow(profile) {
  const row = document.createElement("div");
  row.className = "profile-row";

  const checkCell = document.createElement("label");
  checkCell.className = "check-cell";
  const checkbox = document.createElement("input");
  checkbox.type = "checkbox";
  checkbox.checked = state.selected.has(profile.id);
  checkbox.setAttribute("aria-label", `${profile.name} profilini seç`);
  checkbox.addEventListener("change", () => {
    checkbox.checked ? state.selected.add(profile.id) : state.selected.delete(profile.id);
    updateSelection(visibleProfiles());
  });
  checkCell.appendChild(checkbox);

  const main = document.createElement("div");
  main.className = "profile-main";
  const mark = document.createElement("span");
  mark.className = "family-mark";
  mark.textContent = profile.family === "firefox" ? "FF" : "CH";
  const copy = document.createElement("div");
  copy.className = "profile-copy";
  const name = document.createElement("div");
  name.className = "profile-name";
  name.textContent = profile.name;
  const path = document.createElement("div");
  path.className = "profile-path";
  path.textContent = profile.path;
  path.title = profile.path;
  copy.append(name, path);
  main.append(mark, copy);

  const browserSelect = createBrowserSelect(profile);
  const status = document.createElement("span");
  status.className = "status" + (profile.locked ? " locked" : "");
  status.textContent = profile.locked ? "Kullanımda" : "Hazır";

  const detailButton = document.createElement("button");
  detailButton.className = "button row-button";
  detailButton.textContent = "Detay";
  detailButton.addEventListener("click", () => showDetails(profile, detailButton));

  const openButton = document.createElement("button");
  openButton.className = "button row-button";
  openButton.textContent = "Aç";
  openButton.disabled = browserSelect.disabled;
  openButton.addEventListener("click", () => launchOne(profile, openButton));

  const actions = document.createElement("div");
  actions.className = "row-actions";
  actions.append(detailButton, openButton);

  row.append(checkCell, main, browserSelect, status, actions);
  return row;
}

function createBrowserSelect(profile) {
  const select = document.createElement("select");
  select.className = "browser-select";
  select.setAttribute("aria-label", `${profile.name} için tarayıcı`);
  const compatible = compatibleBrowsers(profile);
  if (!compatible.length) {
    select.appendChild(new Option("Uyumlu tarayıcı yok", ""));
    select.disabled = true;
    return select;
  }
  for (const browser of compatible) {
    select.appendChild(new Option(browser.name, browser.id, false, state.browserChoices.get(profile.id) === browser.id));
  }
  select.addEventListener("change", () => state.browserChoices.set(profile.id, select.value));
  return select;
}

function compatibleBrowsers(profile) {
  return state.browsers.filter((browser) => browser.family === profile.family);
}

function firstCompatibleBrowser(profile) {
  return compatibleBrowsers(profile)[0];
}

async function launchOne(profile, button) {
  const browserId = state.browserChoices.get(profile.id);
  if (!browserId) {
    toast(`${profile.name} için uyumlu tarayıcı bulunamadı.`, true);
    return false;
  }
  const oldText = button?.textContent;
  if (button) {
    button.disabled = true;
    button.textContent = "...";
  }
  try {
    const result = await backend().LaunchProfile(profile, browserId, byId("isolatedToggle").checked);
    if (button) toast(`${result.profile}, ${result.browser} ile açıldı.`);
    return true;
  } catch (error) {
    toast(String(error), true);
    return false;
  } finally {
    if (button) {
      button.disabled = false;
      button.textContent = oldText;
    }
  }
}

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes} B`;
  const units = ["KB", "MB", "GB"];
  let value = bytes / 1024;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) { value /= 1024; unit++; }
  return `${value.toFixed(value < 10 ? 1 : 0)} ${units[unit]}`;
}

function detailSection(title, content) {
  const section = document.createElement("section");
  const heading = document.createElement("h3");
  heading.textContent = title;
  section.append(heading, content);
  return section;
}

function detailList(items) {
  const list = document.createElement("ul");
  for (const item of items) {
    const li = document.createElement("li");
    li.textContent = item;
    list.appendChild(li);
  }
  return list;
}

function detailText(text) {
  const p = document.createElement("p");
  p.textContent = text;
  return p;
}

function renderDetails(details) {
  const body = byId("detailBody");
  body.replaceChildren();

  if (details.notes?.length) {
    const notes = detailList(details.notes);
    notes.className = "detail-notes";
    body.appendChild(notes);
  }
  if (details.fields?.length) {
    const grid = document.createElement("dl");
    for (const field of details.fields) {
      const term = document.createElement("dt");
      term.textContent = field.label;
      const value = document.createElement("dd");
      value.textContent = field.value;
      grid.append(term, value);
    }
    body.appendChild(detailSection("Genel", grid));
  }
  const history = details.history || [];
  if (history.length) {
    body.appendChild(detailSection(
      `Geçmişte girilen siteler (${history.length}) · son ziyarete göre`,
      detailList(history.map((site) => `${site.host} · ${site.visits} ziyaret · son: ${site.lastVisit}`)),
    ));
  }
  const sites = details.sites || [];
  body.appendChild(detailSection(
    `Veri saklayan siteler (${sites.length}) · son değişikliğe göre`,
    sites.length ? detailList(sites.map((site) => `${site.origin} · ${site.modified}`)) : detailText("Site verisi bulunamadı."),
  ));
  const extensions = details.extensions || [];
  if (extensions.length) body.appendChild(detailSection(`Uzantılar (${extensions.length})`, detailList(extensions)));
  if (details.files?.length) {
    body.appendChild(detailSection("Dosyalar", detailList(
      details.files.map((file) => `${file.name}${file.isDir ? "/" : ""} · ${formatSize(file.size)} · ${file.modified}`),
    )));
  }
}

async function showDetails(profile, button) {
  const oldText = button.textContent;
  button.disabled = true;
  button.textContent = "...";
  try {
    const details = await backend().InspectProfile(profile);
    byId("detailTitle").textContent = profile.name;
    byId("detailPath").textContent = profile.path;
    renderDetails(details);
    byId("detailDialog").showModal();
  } catch (error) {
    toast(String(error), true);
  } finally {
    button.disabled = false;
    button.textContent = oldText;
  }
}

function updateSelection(visible = visibleProfiles()) {
  const selectedCount = state.selected.size;
  byId("selectionCount").textContent = `${selectedCount} profil seçili`;
  byId("openSelectedButton").disabled = selectedCount === 0;
  const allVisibleSelected = visible.length > 0 && visible.every((profile) => state.selected.has(profile.id));
  const someVisibleSelected = visible.some((profile) => state.selected.has(profile.id));
  byId("selectAll").checked = allVisibleSelected;
  byId("selectAll").indeterminate = someVisibleSelected && !allVisibleSelected;
}

async function openSelected() {
  const button = byId("openSelectedButton");
  const profiles = state.profiles.filter((profile) => state.selected.has(profile.id));
  button.disabled = true;
  let opened = 0;
  for (let index = 0; index < profiles.length; index++) {
    button.textContent = `${index + 1}/${profiles.length} açılıyor`;
    if (await launchOne(profiles[index])) opened++;
  }
  button.textContent = "Seçilenleri aç";
  button.disabled = state.selected.size === 0;
  toast(`${opened}/${profiles.length} profil açıldı.`, opened !== profiles.length);
}

byId("chooseButton").addEventListener("click", chooseDirectory);
byId("scanButton").addEventListener("click", () => scan());
byId("pathInput").addEventListener("keydown", (event) => { if (event.key === "Enter") scan(); });
byId("searchInput").addEventListener("input", (event) => { state.query = event.target.value; render(); });
byId("familyFilter").addEventListener("change", (event) => { state.filter = event.target.value; render(); });
byId("selectAll").addEventListener("change", (event) => {
  for (const profile of visibleProfiles()) {
    event.target.checked ? state.selected.add(profile.id) : state.selected.delete(profile.id);
  }
  render();
});
byId("openSelectedButton").addEventListener("click", openSelected);
byId("detailClose").addEventListener("click", () => byId("detailDialog").close());

window.addEventListener("DOMContentLoaded", async () => {
  await refreshBrowsers();
  if (window.runtime?.OnFileDrop) {
    window.runtime.OnFileDrop((_x, _y, paths) => {
      if (!paths?.[0]) return;
      byId("pathInput").value = paths[0];
      scan(paths[0]);
    }, true);
  }
});
