---
name: Razorpay Blade
colors:
  background: "#ffffff"
  foreground: "#172b4d"
  brand: "#0d94fb"      # Dodger Blue
  muted: "#5e6c84"
  border: "#ebecf0"
  card: "#ffffff"
  accent: "#012652"     # Prussian Blue
  success: "#04db7c"

  dark:
    background: "#070e1c"
    foreground: "#f4f5f7"
    muted: "#97a0af"
    border: "#1c2536"
    card: "#0f172a"
    accent: "#0d94fb"
    sidebar: "#011638"

typography:
  fontFamily:
    sans: "Mulish, Inter, -apple-system, system-ui, sans-serif"
    mono: "JetBrains Mono, SFMono-Regular, Menlo, monospace"
  body:
    fontSize: "14px"
    lineHeight: "1.5"
  heading:
    fontWeight: "700"
    letterSpacing: "-0.01em"

rounded:
  default: "4px"
  md: "6px"
  lg: "12px"
---

# Design System: Razorpay (Blade)

## Overview
Razorpay's design system, Blade, is a "Developer-First Financial Canvas." It is built on the philosophy of absolute consistency between Figma and Code, using a highly tokenized structure to power financial and risk infrastructure.

## Design Philosophy
1. **Precision Infrastructure:** The UI feels like a high-end developer tool—precise, data-dense, and highly functional.
2. **Transparency:** Every transaction, fee, and settlement is presented with extreme clarity.
3. **The "Blade" Edge:** Uses sharp yet subtly rounded corners (`4px`) and thin lines to communicate efficiency and cutting-edge tech.
4. **Developer-Centric:** Prioritizes documentation-like layouts, clear status codes, and code-friendly patterns.

## Colors
- **Dodger Blue (#0D94FB):** The primary brand action color. It is vibrant and modern, representing the "active" state of the economy.
- **Prussian Blue (#012652 / #011638):** Used for deep headers and sidebars to establish a professional, institutional foundation.
- **Functional Grays:** Uses a meticulously defined scale of greys for backgrounds, borders, and secondary text.

## Typography
- **Mulish (Muli):** A minimalist sans-serif that provides a clean, technical, and modern voice.
- **Data Hierarchy:** Prioritizes monospace fonts for transaction IDs and amounts to ensure they look "engineered."

## Components
- **The Dashboard Sidebar:** A deep Prussian Blue column with high-contrast icons and a clear active state.
- **Payment Link Cards:** Clean surfaces with integrated "Copy" actions and a prominent "Razorpay Blue" primary button.
- **Status Pills:** Semantic pills (Success: Green, Pending: Orange, Failed: Red) with subtle background fills and 4px border radius.

## Visual Effects
- **Tokenized Shadows:** Uses a 3-tier elevation system defined by precise elevation tokens.
- **Clean Grids:** Focuses on data-heavy tables with high-contrast headers and subtle row separation.
- **Interactive Tooltips & Panels:** Uses deep navy backgrounds and high-contrast text for helpful contextual information.
