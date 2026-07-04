<script lang="ts">
	import PageHero from '../_components/PageHero.svelte';
	import { changelogEntries, siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Changelog - OpenPost</title>
	<meta name="description" content="Recent OpenPost Cloud and product updates." />
	<link rel="canonical" href={`${siteUrl}/changelog`} />
</svelte:head>

<PageHero
	eyebrow="Changelog"
	title="Recent product movement."
	description="A compact public history for hosted billing, provider work, automation, and the publishing workflow."
	secondaryHref="/pricing"
	secondaryLabel="See plans"
/>

<section class="section-pad">
	<div class="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
		<div class="divide-y rounded-xl border bg-card">
			{#each changelogEntries as entry (entry.title)}
				<article class="grid gap-4 p-6 sm:grid-cols-[8rem_1fr]">
					<p class="font-mono text-sm text-primary">{entry.date}</p>
					<div>
						<h2 class="text-xl font-semibold">{entry.title}</h2>
						<p class="mt-2 text-sm leading-7 text-muted-foreground">{entry.detail}</p>
					</div>
				</article>
			{/each}
		</div>
	</div>
</section>
