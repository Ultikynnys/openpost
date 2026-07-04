<script lang="ts">
	import { page } from '$app/state';
	import { error } from '@sveltejs/kit';
	import { ArrowRight, CheckCircle2 } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import { appUrl, getPlatform, siteUrl } from '../../_marketing';

	const slug = $derived(page.params.slug ?? '');
	const platform = $derived.by(() => {
		const found = getPlatform(slug);
		if (!found) error(404, 'Platform not found');
		return found;
	});
</script>

<svelte:head>
	<title>Schedule {platform.name} posts - OpenPost</title>
	<meta name="description" content={`Schedule ${platform.name} posts with OpenPost.`} />
	<link rel="canonical" href={`${siteUrl}/platforms/${platform.slug}`} />
</svelte:head>

<section class="section-pad border-b">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[0.8fr_1.2fr] lg:items-center">
			<div>
				<div class="flex size-14 items-center justify-center rounded-xl border bg-card">
					<PlatformIcon platform={platform.short} class="size-7" />
				</div>
				<p class="eyebrow mt-6">{platform.status} provider</p>
				<h1 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
					Schedule {platform.name} posts from OpenPost.
				</h1>
				<p class="mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">
					{platform.description}
				</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href={appUrl} size="lg">
						Open app
						<ArrowRight data-icon="inline-end" />
					</Button>
					<Button href="/platforms" variant="outline" size="lg">All platforms</Button>
				</div>
			</div>
			<div class="rounded-xl border bg-card p-6">
				<h2 class="text-xl font-semibold">Current support notes</h2>
				<ul class="mt-5 grid gap-3">
					{#each platform.limits as limit (limit)}
						<li class="flex gap-3 text-sm leading-6 text-muted-foreground">
							<CheckCircle2 class="mt-0.5 size-4 shrink-0 text-primary" />
							<span>{limit}</span>
						</li>
					{/each}
				</ul>
				<p class="mt-6 text-sm leading-6 text-muted-foreground">
					Provider APIs can change, and video/media behavior varies. OpenPost keeps these limits
					visible instead of flattening every network into the same claim.
				</p>
			</div>
		</div>
	</div>
</section>
