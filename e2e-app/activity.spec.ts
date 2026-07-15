import { expect, test } from "@playwright/test";

test("failed delivery details stay secondary to post status", async ({
  page,
}) => {
  await page.route("**/api/v1/auth/me", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        id: "user-1",
        email: "activity@example.com",
        is_admin: false,
        created_at: "2026-07-01T00:00:00Z",
      },
    });
  });
  await page.route("**/api/v1/workspaces", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        { id: "ws-1", name: "Posts E2E", created_at: "2026-07-01T00:00:00Z" },
      ],
    });
  });
  await page.route("**/api/v1/workspaces/ws-1/settings", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: { timezone: "UTC", week_start: 1 },
    });
  });
  await page.route("**/api/v1/posts**", async (route) => {
    await route.fulfill({ contentType: "application/json", json: [] });
  });
  await page.route("**/api/v1/jobs**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        {
          id: "job-1",
          type: "publish_post",
          status: "failed",
          run_at: "2026-07-01T12:00:00Z",
          attempts: 3,
          max_attempts: 3,
          last_error: "Provider rejected the post",
        },
        {
          id: "job-2",
          type: "publish_thread",
          status: "failed",
          run_at: "2026-07-01T12:05:00Z",
          attempts: 3,
          max_attempts: 3,
          last_error: "Account authorization expired",
        },
      ],
    });
  });

  await page.goto("/activity");

  await expect(page.getByRole("heading", { name: "Posts" })).toBeVisible();
  await expect(page.getByRole("tab", { name: "Jobs" })).toHaveCount(0);
  await page.getByRole("tab", { name: /Failed 2/ }).click();

  const details = page.getByText("Technical details for 2 failed deliveries");
  await expect(details).toBeVisible();
  await details.click();
  await expect(page.getByText("Provider rejected the post")).toBeVisible();
  await expect(page.getByText("Account authorization expired")).toBeVisible();
});
