import { expect, test } from "@playwright/test";
import { authenticatePage, createWorkspace, registerUser } from "./helpers";

const tinyPNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
  "base64",
);

test("mobile shell and composer expose touch-first controls without overflow", async ({
  page,
  request,
}) => {
  await page.setViewportSize({ width: 390, height: 844 });
  const unique = Date.now().toString(36);
  const auth = await registerUser(
    request,
    `mobile-shell-${unique}@example.com`,
  );
  await createWorkspace(request, auth.token, "Mobile UX E2E");
  await authenticatePage(page, auth.token);
  await page.route("**/api/v1/accounts?**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        {
          id: "mobile-bluesky",
          platform: "bluesky",
          account_id: "mobile-bluesky",
          account_username: "openpost_mobile",
          is_active: true,
        },
      ],
    });
  });

  await page.goto("/");

  await expect(
    page.getByRole("button", { name: "Toggle Sidebar" }),
  ).toHaveCount(0);
  await expect(page.getByText("OpenPost", { exact: true })).toHaveCount(0);
  const more = page
    .getByRole("navigation", { name: "Primary navigation" })
    .getByRole("button", { name: "More", exact: true });
  await expect(more).toBeVisible();
  await more.click();
  await expect(page.getByRole("menuitem", { name: "Accounts" })).toBeVisible();
  await expect(page.getByRole("menuitem", { name: "Settings" })).toBeVisible();
  await expect(
    page.getByRole("menuitem", { name: /Appearance/ }),
  ).toBeVisible();
  await expect(
    page.getByRole("menuitemcheckbox", { name: "Interface sounds" }),
  ).toBeVisible();
  await expect(page.getByRole("menuitem", { name: /Language/ })).toBeVisible();
  await page.keyboard.press("Escape");

  const controls = page.getByTestId("mobile-composer-controls");
  await expect(controls).toBeVisible();
  const box = await controls.boundingBox();
  expect(box).not.toBeNull();
  expect(box!.x).toBeGreaterThanOrEqual(0);
  expect(box!.x + box!.width).toBeLessThanOrEqual(390);

  const circles = page.locator(
    '[data-testid="mobile-rendition-all"], [data-testid="mobile-rendition-account"]',
  );
  await expect(circles).toHaveCount(2);
  for (const circle of await circles.all()) {
    const circleBox = await circle.boundingBox();
    expect(circleBox).not.toBeNull();
    expect(circleBox!.width).toBe(circleBox!.height);
    expect(circleBox!.width).toBeGreaterThanOrEqual(44);
  }
});

test("attached media actions are visible on touch and subtle on desktop", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const auth = await registerUser(
    request,
    `mobile-media-${unique}@example.com`,
  );
  const workspace = (await createWorkspace(
    request,
    auth.token,
    "Mobile Media E2E",
  )) as { id: string };
  await authenticatePage(page, auth.token);

  await page.goto("/media");
  await page.getByRole("button", { name: "Upload" }).first().click();
  await page.locator("#file-upload").setInputFiles({
    name: "mobile-actions.png",
    mimeType: "image/png",
    buffer: Buffer.concat([tinyPNG, Buffer.from(unique)]),
  });
  await page
    .getByRole("dialog")
    .getByRole("button", { name: "Upload" })
    .click();
  await expect(page.getByText("File uploaded successfully")).toBeVisible();

  const mediaResponse = await request.get(
    `/api/v1/media?workspace_id=${workspace.id}`,
    {
      headers: { Authorization: `Bearer ${auth.token}` },
    },
  );
  expect(mediaResponse.ok()).toBeTruthy();
  const media = (await mediaResponse.json()) as {
    media: Array<{ id: string }>;
  };
  const draftResponse = await request.post("/api/v1/posts", {
    headers: { Authorization: `Bearer ${auth.token}` },
    data: {
      workspace_id: workspace.id,
      content: "Media accessibility check",
      social_account_ids: [],
      media_ids: [media.media[0].id],
    },
  });
  expect(draftResponse.ok()).toBeTruthy();
  const draft = (await draftResponse.json()) as { id: string };

  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(`/posts/${draft.id}`);
  const actions = page.getByTestId("composer-media-actions");
  await expect(actions).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Add alt text" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Remove media" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Add alt text" }).click();
  await expect(page.getByRole("textbox", { name: "Alt text" })).toBeVisible();
  await page.getByRole("button", { name: "Done" }).click();
  await expect(page.getByRole("textbox", { name: "Alt text" })).toHaveCount(0);

  await page.setViewportSize({ width: 1280, height: 800 });
  await expect(actions).toHaveCSS("opacity", "0");
  await actions.locator("xpath=..").hover();
  await expect(actions).toHaveCSS("opacity", "1");
});
