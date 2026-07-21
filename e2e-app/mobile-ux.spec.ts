import { expect, test, type Locator } from "@playwright/test";
import { authenticatePage, createWorkspace, registerUser } from "./helpers";

const tinyPNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
  "base64",
);

async function expectMinimumTouchTarget(locator: Locator, name: string) {
  await expect(locator, `${name} should be visible`).toBeVisible();
  const box = await locator.boundingBox();
  expect(box, `${name} should have a measurable touch target`).not.toBeNull();
  expect(
    box!.width,
    `${name} should be at least 44px wide`,
  ).toBeGreaterThanOrEqual(44);
  expect(
    box!.height,
    `${name} should be at least 44px tall`,
  ).toBeGreaterThanOrEqual(44);
}

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
  const firstWorkspace = `Mobile UX ${unique}`;
  const secondWorkspace = `Client UX ${unique}`;
  await createWorkspace(request, auth.token, firstWorkspace);
  await createWorkspace(request, auth.token, secondWorkspace);
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
    page.getByRole("menuitem", { name: "Accounts" }),
  ).toHaveAttribute("href", "/accounts");
  await expect(
    page.getByRole("menuitem", { name: "Settings" }),
  ).toHaveAttribute("href", "/settings");
  await expect(
    page.getByRole("menuitem", { name: /Appearance/ }),
  ).toBeVisible();
  await expect(
    page.getByRole("menuitemcheckbox", { name: "Interface sounds" }),
  ).toBeVisible();
  await expect(page.getByRole("menuitem", { name: /Language/ })).toBeVisible();
  const workspaceNames = [firstWorkspace, secondWorkspace];
  const workspaceTrigger = page.getByRole("menuitem", {
    name: new RegExp(`(?:${workspaceNames.join("|")}).*Switch workspace`),
  });
  const workspaceTriggerText = await workspaceTrigger.innerText();
  const currentWorkspace = workspaceNames.find((name) =>
    workspaceTriggerText.includes(name),
  );
  expect(currentWorkspace).toBeTruthy();
  const nextWorkspace =
    currentWorkspace === firstWorkspace ? secondWorkspace : firstWorkspace;
  await expect(workspaceTrigger).toHaveAttribute("aria-expanded", "false");
  await workspaceTrigger.click();
  await expect(workspaceTrigger).toHaveAttribute("aria-expanded", "true");
  await expect(page.getByRole("menuitem", { name: "Accounts" })).toBeVisible();
  await page
    .getByRole("group", { name: "Switch workspace" })
    .getByRole("menuitem", { name: new RegExp(nextWorkspace) })
    .click();
  await more.click();
  await expect(
    page.getByRole("menuitem", { name: new RegExp(nextWorkspace) }),
  ).toContainText(nextWorkspace);
  await page.keyboard.press("Escape");

  const controls = page.getByTestId("mobile-composer-controls");
  await expect(controls).toBeVisible();
  const overflow = await controls.evaluate((element) => ({
    clientWidth: element.clientWidth,
    scrollWidth: element.scrollWidth,
    childRightEdges: Array.from(element.querySelectorAll("button")).map(
      (child) => child.getBoundingClientRect().right,
    ),
  }));
  expect(overflow.scrollWidth).toBeLessThanOrEqual(overflow.clientWidth);
  expect(Math.max(...overflow.childRightEdges)).toBeLessThanOrEqual(390);

  await expect(page.getByTestId("desktop-composer-controls")).toHaveCount(0);
  await expect(page.getByTestId("composer-account-control")).toHaveCount(1);
  await expect(
    page.getByRole("button", { name: "Publish", exact: true }),
  ).toHaveCount(1);
  await expect(
    page.getByRole("button", { name: /^Schedule post:/ }),
  ).toHaveCount(1);

  await expectMinimumTouchTarget(
    page.getByTestId("composer-mode-select"),
    "post type selector",
  );
  await expectMinimumTouchTarget(
    page.getByRole("button", { name: "Add post", exact: true }),
    "add post button",
  );
  await expectMinimumTouchTarget(
    page.locator('label:has(input[type="file"][accept="image/*,video/*"])'),
    "media upload label",
  );

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

  await page.setViewportSize({ width: 1280, height: 800 });
  await expect(page.getByTestId("mobile-composer-controls")).toHaveCount(0);
  await expect(page.getByTestId("desktop-composer-controls")).toBeVisible();
  await expect(page.getByTestId("composer-account-control")).toHaveCount(1);
  await expect(
    page.getByRole("button", { name: "Publish", exact: true }),
  ).toHaveCount(1);
  await expect(
    page.getByRole("button", { name: "Schedule", exact: true }),
  ).toHaveCount(1);
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
  await expect(page.getByRole("dialog")).toHaveCount(0, { timeout: 15_000 });

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
  await page.getByRole("button", { name: "Add alt text" }).focus();
  await expect(actions).toHaveCSS("opacity", "1");
  await page.getByRole("button", { name: "Add alt text" }).blur();
  await expect(actions).toHaveCSS("opacity", "0");
  await actions.locator("xpath=..").hover();
  await expect(actions).toHaveCSS("opacity", "1");
});
