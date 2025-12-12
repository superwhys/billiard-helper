const SCREENS = [
  "welcome",
  "login",
  "register",
  "home",
  "create",
  "snooker",
  "snookerDetail",
  "nineball",
  "eightball",
  "history",
  "refs",
];

function $(sel, root = document) {
  return root.querySelector(sel);
}

function $all(sel, root = document) {
  return Array.from(root.querySelectorAll(sel));
}

function clamp(n, min, max) {
  return Math.max(min, Math.min(max, n));
}

function showToast(text) {
  const toast = $("#toast");
  if (!toast) return;
  toast.textContent = text;
  toast.classList.add("is-show");
  clearTimeout(showToast._t);
  showToast._t = setTimeout(() => toast.classList.remove("is-show"), 1400);
}

function setActive(screen) {
  const key = SCREENS.includes(screen) ? screen : "welcome";

  $all(".screen").forEach((el) => {
    el.classList.toggle("is-active", el.dataset.screen === key);
  });

  $all(".tab").forEach((btn) => {
    btn.classList.toggle("is-active", btn.dataset.go === key);
  });

  $all(".rail__item").forEach((btn) => {
    btn.classList.toggle("is-active", btn.dataset.go === key);
  });

  // bottom nav: only inside visible screen
  $all(".screen").forEach((el) => {
    if (el.dataset.screen !== key) return;
    $all(".bottomNav__item", el).forEach((b) => {
      b.classList.toggle("is-active", b.dataset.go === key);
    });
  });

  document.documentElement.dataset.screen = key;
  history.replaceState(null, "", `#${key}`);
}

function getActive() {
  const cur = document.documentElement.dataset.screen || "welcome";
  return SCREENS.includes(cur) ? cur : "welcome";
}

function next(delta) {
  const cur = getActive();
  const idx = SCREENS.indexOf(cur);
  const nextIdx = clamp(idx + delta, 0, SCREENS.length - 1);
  setActive(SCREENS[nextIdx]);
}

function bindNav() {
  document.addEventListener("click", (e) => {
    const go = e.target.closest("[data-go]");
    if (go) {
      setActive(go.dataset.go);
      return;
    }

    const toast = e.target.closest("[data-toast]");
    if (toast) {
      showToast(toast.dataset.toast);
    }
  });
}

function bindHash() {
  const fromHash = () => {
    const h = (location.hash || "").replace(/^#/, "");
    if (h) setActive(h);
    else setActive("welcome");
  };

  window.addEventListener("hashchange", fromHash);
  fromHash();
}

function bindKeys() {
  window.addEventListener("keydown", (e) => {
    if (e.key === "ArrowLeft") next(-1);
    if (e.key === "ArrowRight") next(1);

    if (e.key === "Escape") {
      const modal = $("#helpModal");
      if (modal?.open) modal.close();
    }
  });
}

function bindHelp() {
  const modal = $("#helpModal");
  const btn = $("#btnHelp");
  const close = $("#btnCloseHelp");
  if (!modal || !btn || !close) return;

  btn.addEventListener("click", () => modal.showModal());
  close.addEventListener("click", () => modal.close());
}

function bindTheme() {
  const btn = $("#btnTheme");
  if (!btn) return;
  btn.addEventListener("click", () => {
    const on = document.documentElement.dataset.theme === "alt";
    document.documentElement.dataset.theme = on ? "" : "alt";
    showToast(on ? "已切回默认主题" : "已切换主题（alt）");

    // alt theme: adjust background only
    if (!on) {
      document.documentElement.style.setProperty("--bg0", "#050510");
      document.documentElement.style.setProperty("--bg1", "#0a0620");
    } else {
      document.documentElement.style.removeProperty("--bg0");
      document.documentElement.style.removeProperty("--bg1");
    }
  });
}

bindNav();
bindHash();
bindKeys();
bindHelp();
bindTheme();
