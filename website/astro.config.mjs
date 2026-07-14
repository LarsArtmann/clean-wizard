import { defineConfig, fontProviders } from "astro/config";
import starlight from "@astrojs/starlight";
import sitemap from "@astrojs/sitemap";

import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  site: "https://cleanwizard.lars.software",

  compressHTML: true,

  prefetch: {
    prefetchAll: false,
    defaultStrategy: "hover",
  },

  fonts: [
    {
      provider: fontProviders.google(),
      name: "Space Grotesk",
      cssVariable: "--font-space-grotesk",
      weights: [300, 400, 500, 600, 700],
      styles: ["normal"],
      subsets: ["latin"],
      fallbacks: ["sans-serif"],
    },
    {
      provider: fontProviders.fontsource(),
      name: "JetBrains Mono",
      cssVariable: "--font-jetbrains-mono",
      weights: [400, 500, 600, 700],
      styles: ["normal"],
      subsets: ["latin"],
      fallbacks: ["monospace"],
    },
  ],

  integrations: [
    sitemap(),
    starlight({
      title: "Clean Wizard",
      favicon: "/favicon.svg",
      customCss: ["./src/styles/starlight.css"],
      expressiveCode: {
        themes: ["github-light", "github-dark"],
        frames: {
          showCopyToClipboardButton: true,
        },
      },
      sidebar: [
        {
          label: "Getting Started",
          items: [
            { label: "Installation", slug: "getting-started/installation" },
            { label: "Quick Start", slug: "getting-started/quick-start" },
          ],
        },
        {
          label: "Guides",
          items: [
            { label: "Cleaners", slug: "guides/cleaners" },
            { label: "Preset Modes", slug: "guides/preset-modes" },
            { label: "Configuration", slug: "guides/configuration" },
            { label: "Automation & CI", slug: "guides/automation" },
          ],
        },
        {
          label: "Reference",
          items: [
            { label: "CLI Reference", slug: "cli-reference" },
            {
              label: "Architecture",
              link: "https://github.com/LarsArtmann/clean-wizard/blob/master/ARCHITECTURE.md",
            },
          ],
        },
        {
          label: "Community",
          items: [
            { label: "Changelog", slug: "changelog" },
            { label: "Contributing", slug: "contributing" },
            { label: "Related Tools", slug: "related-tools" },
          ],
        },
      ],
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/LarsArtmann/clean-wizard",
        },
      ],
      head: [
        {
          tag: "meta",
          attrs: {
            name: "description",
            content:
              "Reclaim disk space across macOS and Linux. 13 specialized cleaners, dry-run previews, interactive TUI, and JSON output for automation.",
          },
        },
      ],
    }),
  ],

  vite: {
    plugins: [tailwindcss()],
  },
});
