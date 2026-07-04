<script lang="ts">
	import { Box, Database, GitBranch, Server } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import PageHero from '../_components/PageHero.svelte';
	import { docsUrl, githubUrl, siteUrl } from '../_marketing';

	const items = [
		['Single binary or container', Box],
		['SQLite by default', Database],
		['No Redis queue requirement', Server],
		['Source-open implementation', GitBranch]
	] as const;
</script>

<svelte:head>
	<title>Open source and self-hosting - OpenPost</title>
	<meta
		name="description"
		content="OpenPost is source open, lightweight to self-host, and available as a managed app."
	/>
	<link rel="canonical" href={`${siteUrl}/open-source`} />
</svelte:head>

<PageHero
	eyebrow="Open source"
	title="Ready to use, but never locked away."
	description="OpenPost is source open and remains practical to inspect, run, and extend."
	secondaryHref={docsUrl}
	secondaryLabel="Read docs"
/>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-4 md:grid-cols-4">
			{#each items as item (item[0])}
				{@const Icon = item[1]}
				<div class="rounded-xl border bg-card p-5">
					<Icon class="size-5 text-primary" />
					<p class="mt-4 font-medium">{item[0]}</p>
				</div>
			{/each}
		</div>
		<div class="mt-10 rounded-xl border bg-muted/25 p-6">
			<h2 class="text-2xl font-semibold">Use self-hosting as the trust signal.</h2>
			<p class="mt-3 max-w-3xl text-sm leading-7 text-muted-foreground">
				OpenPost can run with SQLite and local media storage, while the managed app uses
				Postgres, S3-compatible media storage, and Polar billing. This page gives technical
				buyers confidence that the product is inspectable without making self-hosting the main
				story.
			</p>
			<div class="mt-6 flex flex-wrap gap-3">
				<Button href={githubUrl} target="_blank" rel="noreferrer">View source</Button>
				<Button href={docsUrl} variant="outline">Self-hosting docs</Button>
			</div>
		</div>
	</div>
</section>
