<script lang="ts">
	import { Box, Code2, Database, GitBranch, Server, Waypoints } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import PageHero from '../_components/PageHero.svelte';
	import { appUrl, githubUrl, selfHostingDocsUrl, siteUrl } from '../_marketing';

	const licenseUrl = `${githubUrl}/blob/main/LICENSE`;
	const freedoms = [
		['Inspect the code', Code2],
		['Run it yourself', Server],
		['Modify it', GitBranch],
		['Self-host it', Box]
	] as const;
</script>

<svelte:head>
	<title>Open source and self-hosting - OpenPost</title>
	<meta
		name="description"
		content="OpenPost is open source under AGPL-3.0-only, lightweight to self-host, and available as a managed app."
	/>
	<link rel="canonical" href={`${siteUrl}/open-source`} />
</svelte:head>

<PageHero
	eyebrow="Open source"
	title="Your publishing stack should never be a black box."
	description="OpenPost is open source under AGPL-3.0-only. Inspect it, run it, modify it, and self-host it on infrastructure you control."
	primaryHref={githubUrl}
	primaryLabel="View source"
	secondaryHref={selfHostingDocsUrl}
	secondaryLabel="Self-hosting docs"
/>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each freedoms as item (item[0])}
				{@const Icon = item[1]}
				<div class="rounded-xl border bg-card p-5">
					<Icon class="size-5 text-primary" />
					<p class="mt-4 font-medium">{item[0]}</p>
				</div>
			{/each}
		</div>

		<div class="mt-12 grid gap-10 border-t pt-12 lg:grid-cols-[0.8fr_1.2fr] lg:gap-16">
			<div>
				<p class="eyebrow">AGPL-3.0-only</p>
				<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-4xl">
					Open code, including hosted modifications.
				</h2>
			</div>
			<div class="space-y-5 text-base leading-7 text-muted-foreground">
				<p>
					The GNU Affero General Public License keeps OpenPost and its modified versions open. If
					someone offers a modified version over a network, the people using it must be able to get
					the corresponding source code.
				</p>
				<p>
					That requirement applies to the software, not to posts, media, account data, or other
					content you put into OpenPost.
				</p>
				<Button href={licenseUrl} target="_blank" rel="noreferrer" variant="outline">
					Read the licence
				</Button>
			</div>
		</div>

		<div
			class="mt-12 grid gap-6 rounded-xl border bg-muted/25 p-6 lg:grid-cols-[1fr_auto] lg:items-center lg:p-8"
		>
			<div class="flex gap-4">
				<div
					class="mt-0.5 hidden size-10 shrink-0 items-center justify-center rounded-lg border bg-background sm:flex"
				>
					<Waypoints class="size-5 text-primary" />
				</div>
				<div>
					<h2 class="text-xl font-semibold">Prefer a managed service?</h2>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
						OpenPost also runs a hosted app for people who want updates, infrastructure, and
						operations handled for them. The managed option and self-hosted project share the same
						open-source foundation.
					</p>
				</div>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={appUrl}>Use the managed app</Button>
				<Button href={selfHostingDocsUrl} variant="outline">Run it yourself</Button>
			</div>
		</div>

		<div class="mt-8 flex flex-wrap gap-x-6 gap-y-2 font-mono text-xs text-muted-foreground">
			<span class="inline-flex items-center gap-2"
				><Database class="size-3.5" /> SQLite by default</span
			>
			<span class="inline-flex items-center gap-2"
				><Box class="size-3.5" /> Single binary or container</span
			>
			<span class="inline-flex items-center gap-2"
				><Server class="size-3.5" /> No Redis required</span
			>
		</div>
	</div>
</section>
