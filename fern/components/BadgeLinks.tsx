// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

/**
 * Badge links for GitHub, License, PyPI, etc.
 * Uses a flex wrapper to display badges horizontally and hides Fern's
 * external-link icon that otherwise stacks under each badge image.
 * Requires the `.badge-links` CSS rule from main.css.
 */
export type BadgeItem = {
  href?: string;
  src: string;
  alt: string;
};

export function BadgeLinks({ badges }: { badges: BadgeItem[] }) {
  return (
    <div
      className="badge-links"
      style={{ display: "flex", gap: "8px", flexWrap: "wrap" }}
    >
      {badges.map((b) =>
        b.href ? (
          <a key={b.src} href={b.href} target="_blank" rel="noreferrer">
            <img src={b.src} alt={b.alt} />
          </a>
        ) : (
          <span key={b.src} style={{ display: "inline-flex" }}>
            <img src={b.src} alt={b.alt} />
          </span>
        )
      )}
    </div>
  );
}
