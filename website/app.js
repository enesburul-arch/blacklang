const sidebar = document.querySelector(".sidebar");
const toggle = document.querySelector(".menu-toggle");
const navLinks = Array.from(document.querySelectorAll("nav a"));
const sections = navLinks
  .map((link) => document.querySelector(link.getAttribute("href")))
  .filter(Boolean);
const search = document.querySelector("#navSearch");

toggle?.addEventListener("click", () => {
  document.body.classList.toggle("sidebar-open");
});

navLinks.forEach((link) => {
  link.addEventListener("click", () => {
    document.body.classList.remove("sidebar-open");
    setActiveNav(link.getAttribute("href"));
  });
});

function setActiveNav(href) {
  navLinks.forEach((link) => {
    link.classList.toggle("active", link.getAttribute("href") === href);
  });
}

function updateActiveSection() {
  let active = sections[0];
  const marker = Math.max(24, window.innerHeight * 0.2);

  for (const section of sections) {
    if (section.getBoundingClientRect().top > marker) break;
    active = section;
  }

  if (window.scrollY + window.innerHeight >= document.documentElement.scrollHeight - 2) {
    active = sections[sections.length - 1];
  }

  if (active) setActiveNav(`#${active.id}`);
}

function updateActiveHash() {
  if (navLinks.some((link) => link.getAttribute("href") === window.location.hash)) {
    setActiveNav(window.location.hash);
  } else {
    updateActiveSection();
  }
}

let scrollUpdatePending = false;
function scheduleActiveSection() {
  if (scrollUpdatePending) return;
  scrollUpdatePending = true;
  window.requestAnimationFrame(() => {
    scrollUpdatePending = false;
    updateActiveSection();
  });
}

window.addEventListener("scroll", scheduleActiveSection, { passive: true });
window.addEventListener("resize", scheduleActiveSection);
window.addEventListener("hashchange", updateActiveHash);
window.addEventListener("load", scheduleActiveSection);
updateActiveHash();

search?.addEventListener("input", (event) => {
  const term = event.target.value.trim().toLowerCase();

  navLinks.forEach((link) => {
    const text = link.textContent.toLowerCase();
    link.hidden = term.length > 0 && !text.includes(term);
  });
});

document.querySelectorAll(".code-panel").forEach((panel) => {
  const title = panel.querySelector(".code-title");
  const code = panel.querySelector("code");

  if (!title || !code) return;

  const button = document.createElement("button");
  button.type = "button";
  button.className = "copy-button";
  button.textContent = "Kopyala";
  button.addEventListener("click", async () => {
    await navigator.clipboard.writeText(code.textContent);
    button.textContent = "Kopyalandı";
    window.setTimeout(() => {
      button.textContent = "Kopyala";
    }, 1400);
  });

  title.appendChild(button);
});
