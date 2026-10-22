export const siteConfig = {
  name: "Lampo",
  strapline: "LaTeX, now in real time",
  description:
    "Instant LaTeX previews in your browser. Keep your editor, see every save, export the finished PDF.",
  repoUrl: "https://github.com/amxv/lampo",
  logoHref: "/favicon.svg",
  accentColor: "#ae6539",
  accentColorDark: "#f5c85c",
  footerSections: [
    {
      title: "Lampo",
      text:
        "A little lightning for your LaTeX. Open source, local-first, and made for the joy of writing."
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
  "Get started",
  "Guides",
  "Reference"
] as const;

export const primaryNav = [
  { href: "/docs", label: "Guides" },
  { href: siteConfig.repoUrl, label: "GitHub", external: true }
];
