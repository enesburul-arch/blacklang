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
  });
});

const observer = new IntersectionObserver(
  (entries) => {
    const visible = entries
      .filter((entry) => entry.isIntersecting)
      .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];

    if (!visible) return;

    navLinks.forEach((link) => {
      link.classList.toggle("active", link.getAttribute("href") === `#${visible.target.id}`);
    });
  },
  {
    rootMargin: "-20% 0px -65% 0px",
    threshold: [0.1, 0.25, 0.5],
  },
);

sections.forEach((section) => observer.observe(section));

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
