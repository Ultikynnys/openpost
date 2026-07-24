<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import { Input } from '$lib/components/ui/input';
	import { uploadMediaFile } from '$lib/media-upload-client';
	import { saveStudioBrandKit } from '../api';
	import type {
		StudioBrandAsset,
		StudioBrandColor,
		StudioBrandFont,
		StudioBrandKit,
		StudioBrandTextStyle
	} from '../types';
	import LoaderIcon from 'lucide-svelte/icons/loader-2';
	import PlusIcon from 'lucide-svelte/icons/plus';
	import TrashIcon from 'lucide-svelte/icons/trash-2';
	import UploadIcon from 'lucide-svelte/icons/upload';
	import { m } from '$lib/paraglide/messages';

	let {
		kit,
		onSaved
	}: {
		kit: StudioBrandKit;
		onSaved: (kit: StudioBrandKit) => void;
	} = $props();

	let initialized = false;
	let name = $state('');
	let colors = $state.raw<StudioBrandColor[]>([]);
	let backgrounds = $state.raw<string[]>([]);
	let textStyles = $state.raw<StudioBrandTextStyle[]>([]);
	let assets = $state.raw<StudioBrandAsset[]>([]);
	let fonts = $state.raw<StudioBrandFont[]>([]);
	let saving = $state(false);
	let uploadingAsset = $state(false);
	let uploadingFont = $state(false);
	let error = $state('');
	let success = $state('');
	let fontFamily = $state('');
	let fontWeight = $state(400);
	let fontStyle = $state<'normal' | 'italic'>('normal');
	let fontLicenseAcknowledged = $state(false);

	function initializeEditor() {
		if (initialized) return;
		initialized = true;
		name = kit.name || m.brand_default_name();
		colors = structuredClone($state.snapshot(kit.colors));
		backgrounds = [...kit.backgrounds];
		textStyles = structuredClone($state.snapshot(kit.text_styles));
		assets = structuredClone($state.snapshot(kit.assets));
		fonts = structuredClone($state.snapshot(kit.fonts));
	}

	function updateColor(index: number, field: 'name' | 'value', value: string) {
		colors = colors.map((color, itemIndex) =>
			itemIndex === index ? { ...color, [field]: value } : color
		);
	}

	function updateBackground(index: number, value: string) {
		backgrounds = backgrounds.map((background, itemIndex) =>
			itemIndex === index ? value : background
		);
	}

	function updateTextStyle(
		index: number,
		field: keyof StudioBrandTextStyle,
		value: string | number
	) {
		textStyles = textStyles.map((style, itemIndex) =>
			itemIndex === index ? { ...style, [field]: value } : style
		);
	}

	function addTextStyle() {
		textStyles = [
			...textStyles,
			{
				id: crypto.randomUUID(),
				name: m.brand_text_style_default({ number: textStyles.length + 1 }),
				font_family: fonts[0]?.family || 'Geist Variable',
				font_asset_id: fonts[0]?.media_id,
				font_weight: 700,
				font_style: 'normal',
				font_size: 64,
				color: colors[0]?.value || '#171717',
				line_height: 1.1,
				letter_spacing: 0
			}
		];
	}

	async function uploadBrandAsset(file: File | undefined) {
		if (!file) return;
		uploadingAsset = true;
		error = '';
		try {
			const uploaded = await uploadMediaFile({
				workspaceId: kit.workspace_id,
				file,
				source: 'upload',
				assetKind: 'brand_asset'
			});
			assets = [
				...assets,
				{
					id: crypto.randomUUID(),
					media_id: uploaded.id,
					role: assets.length === 0 ? 'primary_logo' : 'secondary_logo',
					name: file.name.replace(/\.[^.]+$/, '')
				}
			];
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.brand_asset_upload_failed();
		} finally {
			uploadingAsset = false;
		}
	}

	async function uploadBrandFont(file: File | undefined) {
		if (!file) return;
		if (!fontFamily.trim()) {
			error = m.brand_font_family_required();
			return;
		}
		if (!fontLicenseAcknowledged) {
			error = m.brand_font_license_required();
			return;
		}
		uploadingFont = true;
		error = '';
		let objectURL = '';
		try {
			objectURL = URL.createObjectURL(file);
			const previewFamily = `OpenPostFontCheck-${crypto.randomUUID()}`;
			const face = new FontFace(previewFamily, `url("${objectURL}") format("woff2")`, {
				weight: String(fontWeight),
				style: fontStyle
			});
			await face.load();
			const uploaded = await uploadMediaFile({
				workspaceId: kit.workspace_id,
				file: new File([file], file.name, { type: 'font/woff2' }),
				source: 'upload',
				assetKind: 'brand_font'
			});
			fonts = [
				...fonts,
				{
					id: crypto.randomUUID(),
					media_id: uploaded.id,
					family: fontFamily.trim(),
					weight: fontWeight,
					style: fontStyle,
					license_acknowledged: true
				}
			];
			fontFamily = '';
			fontLicenseAcknowledged = false;
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.brand_font_upload_failed();
		} finally {
			if (objectURL) URL.revokeObjectURL(objectURL);
			uploadingFont = false;
		}
	}

	async function save() {
		saving = true;
		error = '';
		success = '';
		try {
			const saved = await saveStudioBrandKit({
				workspace_id: kit.workspace_id,
				name,
				colors,
				text_styles: textStyles,
				backgrounds,
				assets,
				fonts
			});
			onSaved(saved);
			success = m.brand_saved();
		} catch (cause) {
			error = cause instanceof Error ? cause.message : m.brand_save_failed();
		} finally {
			saving = false;
		}
	}
</script>

<div class="space-y-5" {@attach initializeEditor}>
	<div class="flex flex-col gap-3 rounded-xl border bg-card p-4 sm:flex-row sm:items-end">
		<label class="min-w-0 flex-1 text-sm font-medium">
			{m.brand_kit_name()}
			<Input bind:value={name} class="mt-1" maxlength={120} />
		</label>
		<Button onclick={save} disabled={saving || !kit.can_edit}>
			{#if saving}<LoaderIcon class="animate-spin" />{/if}
			{m.brand_save_kit()}
		</Button>
	</div>

	<div class="grid gap-5 lg:grid-cols-2">
		<section class="space-y-4 rounded-xl border bg-card p-4">
			<div>
				<h2 class="font-semibold">{m.brand_colors_backgrounds()}</h2>
				<p class="mt-1 text-sm text-muted-foreground">{m.brand_description()}</p>
			</div>
			<div class="space-y-2">
				{#each colors as color, index (`${index}-${color.name}`)}
					<div class="grid grid-cols-[2.75rem_1fr_8rem_2.75rem] gap-2">
						<input
							type="color"
							value={color.value}
							aria-label={m.brand_choose_color({ name: color.name || m.studio_brand() })}
							class="h-10 w-11 rounded border bg-background p-1"
							oninput={(event) => updateColor(index, 'value', event.currentTarget.value)}
						/>
						<Input
							value={color.name}
							placeholder={m.brand_color_name()}
							oninput={(event) => updateColor(index, 'name', event.currentTarget.value)}
						/>
						<Input
							value={color.value}
							placeholder="#f97316"
							oninput={(event) => updateColor(index, 'value', event.currentTarget.value)}
						/>
						<Button
							variant="ghost"
							size="icon"
							aria-label={m.brand_remove_color()}
							onclick={() => (colors = colors.filter((_, itemIndex) => itemIndex !== index))}
						>
							<TrashIcon />
						</Button>
					</div>
				{/each}
				<Button
					variant="outline"
					size="sm"
					onclick={() =>
						(colors = [
							...colors,
							{ id: crypto.randomUUID(), name: m.brand_new_color(), value: '#f97316' }
						])}
				>
					<PlusIcon />
					{m.brand_add_color()}
				</Button>
			</div>
			<div class="space-y-2 border-t pt-4">
				<h3 class="text-sm font-medium">{m.brand_page_backgrounds()}</h3>
				{#each backgrounds as background, index (`${index}-${background}`)}
					<div class="grid grid-cols-[2.75rem_1fr_2.75rem] gap-2">
						<input
							type="color"
							value={background}
							aria-label={m.brand_choose_background()}
							class="h-10 w-11 rounded border bg-background p-1"
							oninput={(event) => updateBackground(index, event.currentTarget.value)}
						/>
						<Input
							value={background}
							oninput={(event) => updateBackground(index, event.currentTarget.value)}
						/>
						<Button
							variant="ghost"
							size="icon"
							aria-label={m.brand_remove_background()}
							onclick={() =>
								(backgrounds = backgrounds.filter((_, itemIndex) => itemIndex !== index))}
						>
							<TrashIcon />
						</Button>
					</div>
				{/each}
				<Button
					variant="outline"
					size="sm"
					onclick={() => (backgrounds = [...backgrounds, '#ffffff'])}
				>
					<PlusIcon />
					{m.brand_add_background()}
				</Button>
			</div>
		</section>

		<section class="space-y-4 rounded-xl border bg-card p-4">
			<div>
				<h2 class="font-semibold">{m.brand_assets()}</h2>
				<p class="mt-1 text-sm text-muted-foreground">{m.brand_assets_description()}</p>
			</div>
			<label
				class="flex min-h-28 cursor-pointer items-center justify-center rounded-lg border border-dashed text-sm font-medium hover:bg-muted/40"
			>
				{#if uploadingAsset}<LoaderIcon class="mr-2 animate-spin" />{:else}<UploadIcon
						class="mr-2"
					/>{/if}
				{m.brand_upload_asset()}
				<input
					type="file"
					class="sr-only"
					accept="image/png,image/jpeg,image/webp,image/avif"
					disabled={uploadingAsset}
					onchange={(event) => uploadBrandAsset(event.currentTarget.files?.[0])}
				/>
			</label>
			{#each assets as asset, index (asset.id)}
				<div class="grid grid-cols-[1fr_10rem_2.75rem] gap-2">
					<Input
						value={asset.name}
						placeholder={m.brand_asset_name()}
						oninput={(event) =>
							(assets = assets.map((item, itemIndex) =>
								itemIndex === index ? { ...item, name: event.currentTarget.value } : item
							))}
					/>
					<select
						class="h-10 rounded-md border border-input bg-background px-2 text-sm"
						value={asset.role}
						onchange={(event) =>
							(assets = assets.map((item, itemIndex) =>
								itemIndex === index
									? {
											...item,
											role: event.currentTarget.value as StudioBrandAsset['role']
										}
									: item
							))}
					>
						<option value="primary_logo">{m.brand_primary_logo()}</option>
						<option value="secondary_logo">{m.brand_secondary_logo()}</option>
						<option value="mark">{m.brand_mark()}</option>
						<option value="watermark">{m.brand_watermark()}</option>
					</select>
					<Button
						variant="ghost"
						size="icon"
						aria-label={m.brand_remove_asset()}
						onclick={() => (assets = assets.filter((_, itemIndex) => itemIndex !== index))}
					>
						<TrashIcon />
					</Button>
				</div>
			{/each}
		</section>
	</div>

	<section class="space-y-4 rounded-xl border bg-card p-4">
		<div>
			<h2 class="font-semibold">{m.brand_fonts()}</h2>
			<p class="mt-1 text-sm text-muted-foreground">{m.brand_fonts_description()}</p>
		</div>
		<div class="grid gap-3 md:grid-cols-[minmax(12rem,1fr)_7rem_8rem_auto]">
			<Input bind:value={fontFamily} placeholder={m.brand_family_name()} />
			<select
				class="h-10 rounded-md border border-input bg-background px-2 text-sm"
				bind:value={fontWeight}
			>
				<option value={300}>300</option>
				<option value={400}>400</option>
				<option value={500}>500</option>
				<option value={600}>600</option>
				<option value={700}>700</option>
				<option value={800}>800</option>
			</select>
			<select
				class="h-10 rounded-md border border-input bg-background px-2 text-sm"
				bind:value={fontStyle}
			>
				<option value="normal">{m.studio_normal()}</option>
				<option value="italic">{m.studio_italic()}</option>
			</select>
			<label
				class="inline-flex h-10 cursor-pointer items-center justify-center rounded-md border px-3 text-sm font-medium"
			>
				{#if uploadingFont}<LoaderIcon class="mr-2 animate-spin" />{:else}<UploadIcon
						class="mr-2"
					/>{/if}
				{m.brand_upload_woff2()}
				<input
					type="file"
					class="sr-only"
					accept=".woff2,font/woff2"
					disabled={uploadingFont}
					onchange={(event) => uploadBrandFont(event.currentTarget.files?.[0])}
				/>
			</label>
		</div>
		<label class="flex items-start gap-2 text-sm">
			<Checkbox bind:checked={fontLicenseAcknowledged} />
			<span>{m.brand_license_ack()}</span>
		</label>
		<div class="grid gap-2 sm:grid-cols-2">
			{#each fonts as font, index (font.id)}
				<div class="flex items-center justify-between rounded-lg border px-3 py-2">
					<div>
						<p class="text-sm font-medium">{font.family}</p>
						<p class="text-xs text-muted-foreground">{font.weight} · {font.style}</p>
					</div>
					<Button
						variant="ghost"
						size="icon"
						aria-label={m.brand_remove_font()}
						onclick={() => (fonts = fonts.filter((_, itemIndex) => itemIndex !== index))}
					>
						<TrashIcon />
					</Button>
				</div>
			{/each}
		</div>
	</section>

	<section class="space-y-3 rounded-xl border bg-card p-4">
		<div class="flex items-center justify-between gap-3">
			<div>
				<h2 class="font-semibold">{m.studio_text_styles()}</h2>
				<p class="mt-1 text-sm text-muted-foreground">{m.brand_styles_description()}</p>
			</div>
			<Button variant="outline" size="sm" onclick={addTextStyle}
				><PlusIcon /> {m.brand_add_style()}</Button
			>
		</div>
		{#each textStyles as style, index (`${index}-${style.name}`)}
			<div class="grid gap-2 rounded-lg border p-3 sm:grid-cols-2 lg:grid-cols-6">
				<Input
					value={style.name}
					placeholder={m.brand_style_name()}
					oninput={(event) => updateTextStyle(index, 'name', event.currentTarget.value)}
				/>
				<Input
					value={style.font_family}
					placeholder={m.studio_font_family()}
					oninput={(event) => updateTextStyle(index, 'font_family', event.currentTarget.value)}
				/>
				<Input
					type="number"
					min="100"
					max="900"
					step="100"
					value={style.font_weight}
					oninput={(event) =>
						updateTextStyle(index, 'font_weight', event.currentTarget.valueAsNumber)}
				/>
				<Input
					type="number"
					min="6"
					max="512"
					value={style.font_size}
					oninput={(event) =>
						updateTextStyle(index, 'font_size', event.currentTarget.valueAsNumber)}
				/>
				<Input
					value={style.color}
					placeholder="#171717"
					oninput={(event) => updateTextStyle(index, 'color', event.currentTarget.value)}
				/>
				<Button
					variant="ghost"
					class="justify-start text-destructive lg:justify-center"
					onclick={() => (textStyles = textStyles.filter((_, itemIndex) => itemIndex !== index))}
				>
					<TrashIcon />
					{m.brand_remove()}
				</Button>
			</div>
		{/each}
	</section>

	{#if error}<p class="text-sm text-destructive" role="alert">{error}</p>{/if}
	{#if success}<p class="text-sm text-emerald-700 dark:text-emerald-300" role="status">
			{success}
		</p>{/if}
</div>
