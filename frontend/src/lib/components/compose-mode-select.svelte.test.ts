import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ComposeModeSelect from './compose-mode-select.svelte';

describe('ComposeModeSelect', () => {
	it('shows every post type in clear writing and media groups', async () => {
		const onModeChange = vi.fn();
		const screen = await render(ComposeModeSelect, {
			selectedMode: 'short_text',
			onModeChange
		});

		await screen.getByTestId('composer-mode-select').click();

		await expect.element(screen.getByText('Write', { exact: true })).toBeVisible();
		await expect.element(screen.getByText('Media', { exact: true })).toBeVisible();
		for (const label of [
			'Post',
			'Thread',
			'Link',
			'Image',
			'Carousel',
			'Story',
			'Short video',
			'Video'
		]) {
			await expect
				.element(screen.getByRole('option', { name: new RegExp(`^${label}\\b`) }))
				.toBeVisible();
		}
	});

	it('returns restored formats through the existing mode-change contract', async () => {
		const onModeChange = vi.fn();
		const screen = await render(ComposeModeSelect, {
			selectedMode: 'short_text',
			onModeChange
		});

		await screen.getByTestId('composer-mode-select').click();
		await screen.getByTestId('composer-mode-option-carousel').click();

		expect(onModeChange).toHaveBeenCalledOnce();
		expect(onModeChange).toHaveBeenCalledWith('carousel');
	});
});
