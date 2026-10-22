export const siteConfig = {
  name: "Lampo",
  strapline: "Live LaTeX, without the wait",
  description:
    "Lampo is an upcoming lightweight CLI for instant LaTeX previews in the browser and PDF export, designed for agent-driven workflows.",
  repoUrl: "https://github.com/amxv/lampo",
  accentColor: "#b45309",
  accentColorDark: "#fbbf24",
  footerSections: [
    {
      title: "Lampo",
      text:
        "A small, fast live-preview companion for LaTeX projects. Currently in development."
    },
    {
      title: "Project status",
      text:
        "The repository and distribution infrastructure are ready; the preview engine and PDF export are not implemented yet."
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
