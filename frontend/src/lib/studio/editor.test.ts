import { describe, expect, it } from 'vitest';
import { StudioEditor } from './editor.svelte';
import type { StudioDocumentResponse, StudioLayer } from './types';

function layer(id: string, x: number): StudioLayer {
	return {
		id,
		type: 'shape',
		name: id,
		visible: true,
		locked: false,
		opacity: 1,
		transform: {
			x,
			y: 10,
			width: 80,
			height: 80,
			rotation: 0,
			flip_x: false,
			flip_y: false
		},
		shape: {
			kind: 'rectangle',
			fill: '#f97316',
			stroke: '#000000',
			stroke_width: 0,
			radius: 0
		}
	};
}

function response(): StudioDocumentResponse {
	return {
		id: 'design',
		workspace_id: 'workspace',
		created_by_id: 'user',
		revision: 1,
		can_edit: true,
		created_at: '2026-07-25T00:00:00Z',
		updated_at: '2026-07-25T00:00:00Z',
		document: {
			schema_version: 1,
			title: 'Selection test',
			preset_key: 'custom',
			width_px: 1080,
			height_px: 1080,
			brand_kit_revision: 0,
			export_defaults: { format: 'png', quality: 0.92 },
			pages: [
				{
					id: 'page',
					name: 'Page 1',
					background_color: '#ffffff',
					layers: [layer('back', 10), layer('middle', 110), layer('front', 210)]
				}
			]
		}
	};
}

describe('Studio editor layer interactions', () => {
	it('supports native range selection in visual layer order', () => {
		const editor = new StudioEditor();
		editor.load(response());

		editor.selectLayer('front');
		editor.selectLayer('back', 'range');

		expect(new Set(editor.selectedLayerIDs)).toEqual(new Set(['front', 'middle', 'back']));
	});

	it('keeps grouped children nested and deletes a selected group as one unit', () => {
		const editor = new StudioEditor();
		editor.load(response());
		editor.selectLayer('front');
		editor.selectLayer('middle', 'toggle');

		editor.groupSelected();

		const group = editor.selectedLayers[0];
		expect(group.type).toBe('group');
		expect(
			editor.activePage?.layers
				.filter((candidate) => candidate.parent_id === group.id)
				.map((candidate) => candidate.id)
		).toEqual(expect.arrayContaining(['front', 'middle']));

		editor.deleteSelected();

		expect(editor.activePage?.layers.map((candidate) => candidate.id)).toEqual(['back']);
	});

	it('adds persistent pencil and bucket paint layers above the active layer', () => {
		const editor = new StudioEditor();
		editor.load(response());
		editor.selectLayer('front');
		editor.paintColor = '#0ea5e9';
		editor.pencilSize = 8;

		editor.addPencilStroke([
			{ x: 220, y: 30 },
			{ x: 280, y: 70 }
		]);

		const pencil = editor.activePage?.layers.at(-1);
		expect(pencil?.type).toBe('paint');
		expect(pencil?.paint?.kind).toBe('fill');
		expect(pencil?.paint?.color).toBe('#0ea5e9');

		const mask = new Uint8Array(1080 * 1080);
		mask.fill(1, 10 * 1080 + 20, 10 * 1080 + 40);
		editor.addPaintFill(mask);

		const bucket = editor.activePage?.layers.at(-1);
		expect(bucket?.type).toBe('paint');
		expect(bucket?.paint?.kind).toBe('fill');
		expect(bucket?.paint?.spans).toEqual([{ x: 0, y: 0, width: 20 }]);
	});
});
