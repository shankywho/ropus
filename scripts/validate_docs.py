#!/usr/bin/env python3
"""
Real Documentation & Markdown Link Validator for ROPUS
Validates all internal relative Markdown links, ensures referenced target files exist,
and rejects forbidden machine-specific file:// or absolute path URLs.
"""

import os
import re
import sys
from urllib.parse import unquote

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

REQUIRED_DOCS = [
    "README.md",
    "METRICS.md",
    "docs/README.md",
    "docs/buildathon-evidence.md",
    "docs/limitations.md",
    "docs/demo-runbook.md",
    "docs/SUBMISSION_GUIDE.md",
    "docs/ml_quality_report.md",
]

# Regex for Markdown links: [anchor text](target)
LINK_PATTERN = re.compile(r'\[([^\]]+)\]\(([^)]+)\)')

def validate_docs() -> bool:
    print("=" * 70)
    print("ROPUS Documentation & Link Integrity Validator")
    print("=" * 70)

    # 1. Verify required documents exist
    print("\n[STEP 1/3] Checking presence of core Buildathon documentation...")
    missing_required = []
    for doc in REQUIRED_DOCS:
        full_path = os.path.join(REPO_ROOT, doc)
        if not os.path.isfile(full_path):
            print(f"  ✗ MISSING: {doc}")
            missing_required.append(doc)
        else:
            print(f"  ✓ FOUND:   {doc}")

    if missing_required:
        print(f"\n[ERROR] {len(missing_required)} required document(s) missing!")
        return False

    # 2. Discover all Markdown files
    print("\n[STEP 2/3] Scanning repository for Markdown files...")
    md_files = []
    for root, dirs, files in os.walk(REPO_ROOT):
        # Exclude vendor and cache directories
        dirs[:] = [d for d in dirs if d not in {
            "node_modules", ".git", ".system_generated", ".next", ".tanstack",
            "dist", "build", ".output", ".pytest_cache", "__pycache__"
        }]
        for f in files:
            if f.endswith(".md"):
                md_files.append(os.path.join(root, f))

    print(f"  Found {len(md_files)} Markdown files to validate.")

    # 3. Validate links inside all Markdown files
    print("\n[STEP 3/3] Validating internal relative links & checking for forbidden paths...")
    broken_links = []
    forbidden_urls = []
    total_links_checked = 0

    for md_path in md_files:
        rel_md_path = os.path.relpath(md_path, REPO_ROOT)
        with open(md_path, "r", encoding="utf-8", errors="ignore") as f:
            content = f.read()

        links = LINK_PATTERN.findall(content)
        for anchor_text, raw_target in links:
            total_links_checked += 1
            target = raw_target.strip()

            # Ignore empty or pure anchor links (#section)
            if not target or target.startswith("#"):
                continue

            # Ignore external protocols (http, https, mailto)
            if target.startswith("http://") or target.startswith("https://") or target.startswith("mailto:"):
                continue

            # Check for forbidden machine-specific file:// or absolute paths
            if target.startswith("file://") or target.startswith("/Users/") or target.startswith("/home/"):
                forbidden_urls.append((rel_md_path, anchor_text, target, "Forbidden absolute or file:// path"))
                continue

            # Clean target: remove query params and URL fragments
            clean_target = target.split("#")[0].split("?")[0]
            if not clean_target:
                continue

            clean_target = unquote(clean_target)

            # Resolve target path relative to the directory of the current Markdown file
            md_dir = os.path.dirname(md_path)
            resolved_path_rel = os.path.normpath(os.path.join(md_dir, clean_target))
            # Also try resolving relative to repository root
            resolved_path_root = os.path.normpath(os.path.join(REPO_ROOT, clean_target))

            if not os.path.exists(resolved_path_rel) and not os.path.exists(resolved_path_root):
                broken_links.append((rel_md_path, anchor_text, target, resolved_path_rel))

    print(f"  Total Markdown links evaluated: {total_links_checked}")

    has_errors = False
    if forbidden_urls:
        has_errors = True
        print(f"\n[ERROR] Found {len(forbidden_urls)} forbidden machine-specific or file:// link(s):")
        for source_file, anchor, target, reason in forbidden_urls:
            print(f"  • {source_file}: [{anchor}]({target}) -> {reason}")

    if broken_links:
        has_errors = True
        print(f"\n[ERROR] Found {len(broken_links)} broken internal Markdown link(s):")
        for source_file, anchor, target, resolved in broken_links:
            print(f"  • {source_file}: [{anchor}]({target}) -> Target not found at '{os.path.relpath(resolved, REPO_ROOT)}'")

    if not has_errors:
        print("\n" + "=" * 70)
        print("✓ ALL DOCUMENTATION AND RELATIVE LINKS VALIDATED SUCCESSFULLY!")
        print("=" * 70)
        return True
    else:
        print("\n" + "=" * 70)
        print("✗ DOCUMENTATION VALIDATION FAILED!")
        print("=" * 70)
        return False

if __name__ == "__main__":
    success = validate_docs()
    sys.exit(0 if success else 1)
