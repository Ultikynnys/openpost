import { expect, test } from "@playwright/test";

test("marketing index links to the app and documentation", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle(
    /OpenPost - Social publishing for creators and small teams/,
  );
  await expect(
    page.getByRole("heading", {
      name: "The open social publishing workspace without the enterprise bloat.",
    }),
  ).toBeVisible();
  await expect(
    page.getByText("Source-open publishing, ready to use"),
  ).toBeVisible();
  await expect(
    page.getByRole("link", { name: "Start publishing" }).first(),
  ).toHaveAttribute("href", "https://app.openpost.social");
  await expect(
    page.getByRole("link", { name: "Self-host instead" }).first(),
  ).toHaveAttribute("href", "https://docs.openpost.social/self-hosting/");
  await expect(
    page.getByRole("link", { name: "User docs" }).first(),
  ).toHaveAttribute("href", "https://docs.openpost.social/usage/");
  await expect(
    page.getByRole("link", { name: "Self-hosting docs" }).first(),
  ).toHaveAttribute("href", "https://docs.openpost.social/self-hosting/");
  await expect(
    page.getByRole("link", { name: "Developer docs" }).first(),
  ).toHaveAttribute("href", "https://docs.openpost.social/development/");
  await expect(
    page.getByRole("link", { name: "GitHub source" }),
  ).toHaveAttribute("href", "https://github.com/rodrgds/openpost");
});

test("marketing index has no horizontal overflow", async ({ page }) => {
  await page.goto("/");

  const overflow = await page.evaluate(
    () =>
      document.documentElement.scrollWidth -
      document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(1);
});

test("marketing SEO routes expose the current public index", async ({
  request,
}) => {
  const robots = await request.get("/robots.txt");
  expect(robots.ok()).toBeTruthy();
  const robotsText = await robots.text();
  expect(robotsText).toContain("Sitemap: https://openpost.social/sitemap.xml");

  const sitemap = await request.get("/sitemap.xml");
  expect(sitemap.ok()).toBeTruthy();
  const xml = await sitemap.text();
  expect(xml).toContain("<loc>https://openpost.social/</loc>");
  expect(xml).toContain("<loc>https://openpost.social/platforms</loc>");
  expect(xml).toContain("<loc>https://openpost.social/platforms/x</loc>");
  expect(xml).toContain("<loc>https://openpost.social/compare</loc>");
  expect(xml).toContain("<loc>https://openpost.social/compare/buffer</loc>");
  expect(xml).toContain("<loc>https://openpost.social/tools</loc>");
  expect(xml).toContain(
    "<loc>https://openpost.social/tools/multi-platform-character-counter</loc>",
  );
  expect(xml).toContain("<loc>https://openpost.social/security</loc>");
  expect(xml).toContain("<loc>https://openpost.social/privacy</loc>");
  expect(xml).toContain("<loc>https://openpost.social/terms</loc>");
  expect(xml).not.toContain("<loc>https://openpost.social/blog</loc>");
  expect(xml).not.toContain("<loc>https://openpost.social/tips/");
});
