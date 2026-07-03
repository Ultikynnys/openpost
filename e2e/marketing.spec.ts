import { expect, test } from "@playwright/test";

test("marketing index links to the app and documentation", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle(/OpenPost - Open source social publishing/);
  await expect(
    page.getByRole("heading", { name: "OpenPost" }),
  ).toBeVisible();
  await expect(
    page.getByText("Open source social publishing"),
  ).toBeVisible();
  await expect(page.getByRole("link", { name: "Open app" }).first()).toHaveAttribute(
    "href",
    "https://app.openpost.social",
  );
  await expect(page.getByRole("link", { name: "Read docs" }).first()).toHaveAttribute(
    "href",
    "https://docs.openpost.social",
  );
  await expect(
    page.getByRole("link", { name: "Docs", exact: true }),
  ).toHaveAttribute("href", "https://docs.openpost.social");
  await expect(page.getByRole("link", { name: "Source" })).toHaveAttribute(
    "href",
    "https://github.com/rodrgds/openpost",
  );
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

test("marketing SEO routes expose only the current public index", async ({
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
  expect(xml).not.toContain("<loc>https://openpost.social/tools</loc>");
  expect(xml).not.toContain("<loc>https://openpost.social/blog</loc>");
  expect(xml).not.toContain("<loc>https://openpost.social/compare/");
  expect(xml).not.toContain("<loc>https://openpost.social/tips/");
});
