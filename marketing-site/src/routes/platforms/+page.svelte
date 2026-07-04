<script lang="ts">
	import { ArrowRight } from 'lucide-svelte';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import PageHero from '../_components/PageHero.svelte';
	import { platforms, siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Supported platforms - OpenPost</title>
	<meta
		name="description"
		content="OpenPost platform support for X, Mastodon, Bluesky, LinkedIn, Threads, Facebook, Instagram, TikTok, and YouTube."
	/>
	<link rel="canonical" href={`${siteUrl}/platforms`} />
</svelte:head>

<PageHero
	eyebrow="Platforms"
	title="Platform support, described without pretending every network is the same."
	description="OpenPost uses one composer and per-account variants, but provider support still follows each network's real API behavior."
	secondaryHref="/compare"
	secondaryLabel="Compare tools"
/>

<section class="section-pad">
	<div class="mx-auto grid max-w-7xl gap-4 px-4 sm:px-6 md:grid-cols-2 lg:grid-cols-3 lg:px-8">
		{#each platforms as platform (platform.slug)}
			<a href={`/platforms/${platform.slug}`} class="group rounded-xl border bg-card p-5 transition hover:bg-muted/30">
				<div class="flex items-start justify-between gap-4">
					<div class="flex items-center gap-3">
						<div class="flex size-10 items-center justify-center rounded-lg border bg-muted">
							<PlatformIcon platform={platform.short} class="size-5" />
						</div>
						<div>
							<h2 class="font-semibold">{platform.name}</h2>
							<p class="text-xs text-muted-foreground">{platform.status}</p>
						</div>
					</div>
					<ArrowRight class="size-4 text-muted-foreground transition group-hover:translate-x-0.5 group-hover:text-foreground" />
				</div>
				<p class="mt-4 text-sm leading-6 text-muted-foreground">{platform.description}</p>
				<div class="mt-4 flex flex-wrap gap-2">
					<span class="rounded-full border px-2 py-1 text-xs text-muted-foreground">{platform.tag}</span>
				</div>
			</a>
		{/each}
	</div>
</section>
