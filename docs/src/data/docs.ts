export const siteConfig = {
  name: "Lampo",
  strapline: "Live LaTeX, without the wait",
  description:
    "Lampo is a lightweight local CLI for live LaTeX previews in the browser and one-click PDF export, designed for agent-driven workflows.",
  repoUrl: "https://github.com/amxv/lampo",
  accentColor: "#b45309",
  accentColorDark: "#fbbf24",
  footerSections: [
    {
      title: "Lampo",
      text:
        "A small, fast and local-first live-preview companion for LaTeX projects."
    },
    {
      title: "Project status",
      text:
        "The live preview works from source. The first npm and native binary release is still being tested."
    },
    {
      title: "Repository",
      linkPrefix: "Source: ",
      linkHref: "https://github.com/amxv/lampo",
      linkLabel: "github.com/amxv/lampo"
    }
  ]
} as const;

export const docCategories = [
  "Start",
  "Development",
  "Distribution",
  "Reference"
] as const;

export const primaryNav = [
  { href: "/docs", label: "Docs" },
  { href: siteConfig.repoUrl, label: "GitHub", external: true }
];
