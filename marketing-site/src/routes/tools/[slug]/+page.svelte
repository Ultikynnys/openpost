<script lang="ts">
	import { page } from '$app/state';
	import { error } from '@sveltejs/kit';
	import { ArrowRight, CheckCircle2, Copy, Wand2 } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import {
		X_PREMIUM_CHAR_LIMIT,
		publicPlatformLimits,
		platformCharacterLimit,
		type PlatformLimitDefinition
	} from '$lib/platform-limits';
	import { appUrl, getTool, siteUrl } from '../../_marketing';

	const slug = $derived(page.params.slug ?? '');
	const tool = $derived.by(() => {
		const found = getTool(slug);
		if (!found) error(404, 'Tool not found');
		return found;
	});

	let draft = $state(
		'Launch note: OpenPost keeps social publishing focused. Draft once, adapt per platform, preview the destination shape, then schedule into the next workspace slot.'
	);
	let selectedPlatform = $state('x');
	let xPremium = $state(false);
	let formatterMode = $state<'bold' | 'italic' | 'clear'>('bold');
	let timezone = $state('Europe/Lisbon');
	let cadence = $state<'weekday' | 'daily' | 'launch'>('weekday');
	let handleInput = $state('@openpost@mastodon.social');

	const publicLimits = publicPlatformLimits();
	const selectablePlatforms = publicLimits.filter((platform) =>
		['x', 'mastodon', 'bluesky', 'linkedin', 'threads', 'facebook'].includes(platform.key)
	);
	const previewPlatforms = publicLimits.filter((platform) =>
		['x', 'mastodon', 'bluesky', 'linkedin', 'threads'].includes(platform.key)
	);
	const currentLimit = $derived(
		selectedPlatform === 'x' && xPremium
			? X_PREMIUM_CHAR_LIMIT
			: platformCharacterLimit(selectedPlatform)
	);
	const draftLength = $derived(draft.length);
	const remaining = $derived(currentLimit - draftLength);
	const threadParts = $derived.by(() => splitThread(draft, currentLimit));
	const handleResult = $derived.by(() => parseHandle(handleInput));
	const postingSlots = $derived.by(() => buildSlots(timezone, cadence));
	const formattedDraft = $derived.by(() => formatLinkedIn(draft, formatterMode));

	function platformDisplay(platform: PlatformLimitDefinition) {
		if (platform.key === 'x') return xPremium ? 'X Premium' : platform.name;
		return platform.name;
	}

	function limitFor(platform: PlatformLimitDefinition) {
		if (platform.key === 'x' && xPremium) return X_PREMIUM_CHAR_LIMIT;
		return platform.charLimit;
	}

	function splitThread(value: string, limit: number) {
		const text = value.trim().replace(/\s+/g, ' ');
		if (!text) return [];
		const words = text.split(' ');
		const parts: string[] = [];
		let current = '';

		for (const word of words) {
			const next = current ? `${current} ${word}` : word;
			if (next.length <= limit) {
				current = next;
				continue;
			}
			if (current) parts.push(current);
			current = word.length > limit ? word.slice(0, limit) : word;
		}
		if (current) parts.push(current);
		return parts;
	}

	function parseHandle(value: string) {
		const trimmed = value.trim();
		const mastodon = /^@?([a-zA-Z0-9_]+)@([a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$/.exec(trimmed);
		if (mastodon) {
			return {
				valid: true,
				type: 'Mastodon / Fediverse',
				username: mastodon[1],
				host: mastodon[2],
				message: 'Looks like a valid Fediverse handle format.'
			};
		}

		const bluesky = /^@?([a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)+)$/.exec(trimmed);
		if (bluesky) {
			return {
				valid: true,
				type: 'Bluesky handle',
				username: bluesky[1],
				host: bluesky[1].split('.').slice(1).join('.'),
				message: 'Looks like a valid domain-style Bluesky handle.'
			};
		}

		return {
			valid: false,
			type: 'Unknown',
			username: '',
			host: '',
			message: 'Use @name@server.tld for Mastodon or name.bsky.social for Bluesky.'
		};
	}

	function buildSlots(zone: string, mode: typeof cadence) {
		const base =
			mode === 'launch'
				? ['09:15', '12:30', '16:45', '19:10']
				: mode === 'daily'
					? ['08:45', '12:15', '17:30']
					: ['09:30', '13:00', '16:30'];
		return base.map((time, index) => ({
			time,
			label: ['Primary', 'Midday', 'Follow-up', 'Evening'][index],
			zone
		}));
	}

	function mapText(value: string, style: 'bold' | 'italic') {
		const upperBase = style === 'bold' ? 0x1d400 : 0x1d434;
		const lowerBase = style === 'bold' ? 0x1d41a : 0x1d44e;
		const digitBase = style === 'bold' ? 0x1d7ce : null;
		return [...value]
			.map((char) => {
				const code = char.charCodeAt(0);
				if (code >= 65 && code <= 90) return String.fromCodePoint(upperBase + code - 65);
				if (code >= 97 && code <= 122) return String.fromCodePoint(lowerBase + code - 97);
				if (digitBase && code >= 48 && code <= 57) return String.fromCodePoint(digitBase + code - 48);
				return char;
			})
			.join('');
	}

	function formatLinkedIn(value: string, mode: typeof formatterMode) {
		if (mode === 'clear') return value.replace(/[ \t]+$/gm, '').replace(/\n{3,}/g, '\n\n');
		return mapText(value, mode);
	}

	async function copyText(value: string) {
		await navigator.clipboard?.writeText(value);
	}
</script>

<svelte:head>
	<title>{tool.name} - OpenPost</title>
	<meta name="description" content={tool.description} />
	<link rel="canonical" href={`${siteUrl}/tools/${tool.slug}`} />
</svelte:head>

<section class="section-pad border-b">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[0.85fr_1.15fr] lg:items-center">
			<div>
				<p class="eyebrow">Free tool</p>
				<h1 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
					{tool.name}
				</h1>
				<p class="mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">{tool.description}</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href={appUrl} size="lg">
						Open the app
						<ArrowRight data-icon="inline-end" />
					</Button>
					<Button href="/tools" variant="outline" size="lg">All tools</Button>
				</div>
			</div>
			<div class="rounded-xl border bg-card p-5">
				<label for="tool-input" class="text-sm font-medium">Draft</label>
				<textarea
					id="tool-input"
					bind:value={draft}
					class="mt-3 min-h-44 w-full rounded-md border bg-background p-4 text-sm leading-6"
					placeholder="Paste your social post here..."
				></textarea>

				{#if tool.slug === 'multi-platform-character-counter'}
					<div class="mt-4 flex flex-wrap items-center gap-3">
						<label class="inline-flex items-center gap-2 text-sm text-muted-foreground">
							<input type="checkbox" bind:checked={xPremium} class="size-4 rounded border" />
							Use X Premium longer-post limit
						</label>
					</div>
					<div class="mt-4 grid gap-3 sm:grid-cols-2">
						{#each selectablePlatforms as platform (platform.key)}
							{@const limit = limitFor(platform)}
							<div class="rounded-md border bg-muted/25 p-3">
								<div class="flex items-center justify-between gap-3">
									<span class="inline-flex items-center gap-2 text-sm font-medium">
										<PlatformIcon platform={platform.key} class="size-4" />
										{platformDisplay(platform)}
									</span>
									<span class="font-mono text-xs {draftLength > limit ? 'text-red-500' : 'text-muted-foreground'}">
										{draftLength}/{limit.toLocaleString()}
									</span>
								</div>
								<div class="mt-3 h-2 overflow-hidden rounded-full bg-muted">
									<div
										class="h-full rounded-full {draftLength > limit ? 'bg-red-500' : 'bg-primary'}"
										style={`width: ${Math.min(100, (draftLength / limit) * 100)}%`}
									></div>
								</div>
							</div>
						{/each}
					</div>
				{:else if tool.slug === 'thread-splitter'}
					<div class="mt-4 flex flex-wrap items-center gap-3">
						<select bind:value={selectedPlatform} class="rounded-md border bg-background px-3 py-2 text-sm">
							{#each selectablePlatforms as platform (platform.key)}
								<option value={platform.key}>{platform.name}</option>
							{/each}
						</select>
						{#if selectedPlatform === 'x'}
							<label class="inline-flex items-center gap-2 text-sm text-muted-foreground">
								<input type="checkbox" bind:checked={xPremium} class="size-4 rounded border" />
								X Premium
							</label>
						{/if}
						<span class="font-mono text-xs text-muted-foreground">
							{threadParts.length || 0} part{threadParts.length === 1 ? '' : 's'} at {currentLimit.toLocaleString()} chars
						</span>
					</div>
					<div class="mt-4 grid gap-3">
						{#each threadParts as part, index (`part-${index}`)}
							<div class="rounded-md border bg-muted/25 p-3">
								<div class="flex items-center justify-between gap-3">
									<span class="font-mono text-xs text-muted-foreground">Post {index + 1}</span>
									<Button type="button" size="sm" variant="ghost" onclick={() => copyText(part)}>
										<Copy data-icon="inline-start" />
										Copy
									</Button>
								</div>
								<p class="mt-2 text-sm leading-6">{part}</p>
							</div>
						{/each}
					</div>
				{:else if tool.slug === 'linkedin-text-formatter'}
					<div class="mt-4 flex flex-wrap gap-2">
						{#each [
							['bold', 'Bold'],
							['italic', 'Italic'],
							['clear', 'Clean spacing']
						] as option (option[0])}
							<Button
								type="button"
								variant={formatterMode === option[0] ? 'default' : 'outline'}
								size="sm"
								onclick={() => (formatterMode = option[0] as typeof formatterMode)}
							>
								<Wand2 data-icon="inline-start" />
								{option[1]}
							</Button>
						{/each}
					</div>
					<div class="mt-4 rounded-md border bg-muted/25 p-4">
						<div class="mb-3 flex items-center justify-between gap-3">
							<span class="text-sm font-medium">Formatted copy</span>
							<Button type="button" size="sm" variant="ghost" onclick={() => copyText(formattedDraft)}>
								<Copy data-icon="inline-start" />
								Copy
							</Button>
						</div>
						<p class="whitespace-pre-wrap text-sm leading-6">{formattedDraft}</p>
					</div>
				{:else if tool.slug === 'best-time-to-post-calculator'}
					<div class="mt-4 grid gap-3 sm:grid-cols-2">
						<label class="grid gap-2 text-sm">
							<span class="font-medium">Timezone</span>
							<select bind:value={timezone} class="rounded-md border bg-background px-3 py-2">
								<option>Europe/Lisbon</option>
								<option>UTC</option>
								<option>America/New_York</option>
								<option>America/Los_Angeles</option>
								<option>Europe/London</option>
							</select>
						</label>
						<label class="grid gap-2 text-sm">
							<span class="font-medium">Cadence</span>
							<select bind:value={cadence} class="rounded-md border bg-background px-3 py-2">
								<option value="weekday">Weekday publishing</option>
								<option value="daily">Daily publishing</option>
								<option value="launch">Launch day</option>
							</select>
						</label>
					</div>
					<div class="mt-4 grid gap-3 sm:grid-cols-3">
						{#each postingSlots as slot (slot.time)}
							<div class="rounded-md border bg-muted/25 p-4">
								<p class="font-mono text-lg font-semibold">{slot.time}</p>
								<p class="mt-1 text-xs text-muted-foreground">{slot.label} · {slot.zone}</p>
							</div>
						{/each}
					</div>
				{:else if tool.slug === 'fediverse-handle-checker'}
					<label for="handle-input" class="mt-4 block text-sm font-medium">Handle</label>
					<input
						id="handle-input"
						bind:value={handleInput}
						class="mt-2 w-full rounded-md border bg-background px-3 py-2 text-sm"
						placeholder="@name@mastodon.social or name.bsky.social"
					/>
					<div class="mt-4 rounded-md border bg-muted/25 p-4">
						<div class="flex items-center gap-2">
							<CheckCircle2 class="size-4 {handleResult.valid ? 'text-primary' : 'text-muted-foreground'}" />
							<span class="font-medium">{handleResult.type}</span>
						</div>
						<p class="mt-2 text-sm leading-6 text-muted-foreground">{handleResult.message}</p>
						{#if handleResult.valid}
							<p class="mt-3 font-mono text-xs text-muted-foreground">
								username={handleResult.username} host={handleResult.host}
							</p>
						{/if}
					</div>
				{:else}
					<div class="mt-4 flex flex-wrap items-center gap-3">
						<select bind:value={selectedPlatform} class="rounded-md border bg-background px-3 py-2 text-sm">
							{#each previewPlatforms as platform (platform.key)}
								<option value={platform.key}>{platform.name}</option>
							{/each}
						</select>
						{#if selectedPlatform === 'x'}
							<label class="inline-flex items-center gap-2 text-sm text-muted-foreground">
								<input type="checkbox" bind:checked={xPremium} class="size-4 rounded border" />
								X Premium
							</label>
						{/if}
						<span class="font-mono text-xs {remaining < 0 ? 'text-red-500' : 'text-muted-foreground'}">
							{remaining.toLocaleString()} remaining
						</span>
					</div>
					<div class="mt-4 rounded-xl border bg-background p-4">
						<div class="flex items-center gap-3">
							<div class="grid size-10 place-items-center rounded-full border bg-muted">
								<PlatformIcon platform={selectedPlatform} class="size-5" />
							</div>
							<div>
								<p class="text-sm font-semibold">OpenPost</p>
								<p class="text-xs text-muted-foreground">@openpost</p>
							</div>
						</div>
						<p class="mt-4 whitespace-pre-wrap text-sm leading-6">{draft}</p>
						<div class="mt-4 aspect-video rounded-lg border bg-muted/35"></div>
					</div>
				{/if}

				<p class="mt-4 text-xs leading-5 text-muted-foreground">
					The full app adds connected accounts, media uploads, variants, workspace schedules,
					per-account destinations, and queue state.
				</p>
			</div>
		</div>
	</div>
</section>
