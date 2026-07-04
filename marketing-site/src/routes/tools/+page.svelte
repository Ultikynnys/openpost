<script lang="ts">
	import { ArrowRight } from 'lucide-svelte';
	import PageHero from '../_components/PageHero.svelte';
	import { siteUrl, tools } from '../_marketing';
</script>

<svelte:head>
	<title>Free social media tools - OpenPost</title>
	<meta
		name="description"
		content="Free social media tools from OpenPost: character counter, preview generator, thread splitter, handle checker, and scheduling helpers."
	/>
	<link rel="canonical" href={`${siteUrl}/tools`} />
</svelte:head>

<PageHero
	eyebrow="Free tools"
	title="Small tools pulled from the same publishing logic."
	description="Use these as lightweight utilities before you need the full OpenPost Cloud workspace."
	secondaryHref="/pricing"
	secondaryLabel="See Cloud plans"
/>

<section class="section-pad">
	<div class="mx-auto grid max-w-7xl gap-4 px-4 sm:px-6 md:grid-cols-2 lg:grid-cols-3 lg:px-8">
		{#each tools as tool (tool.slug)}
			{@const Icon = tool.icon}
			<a href={`/tools/${tool.slug}`} class="group rounded-xl border bg-card p-5 transition hover:bg-muted/30">
				<div class="flex items-start justify-between gap-4">
					<div class="flex size-10 items-center justify-center rounded-lg border bg-muted">
						<Icon class="size-5 text-primary" />
					</div>
					<ArrowRight class="size-4 text-muted-foreground transition group-hover:translate-x-0.5 group-hover:text-foreground" />
				</div>
				<h2 class="mt-5 font-semibold">{tool.name}</h2>
				<p class="mt-3 text-sm leading-6 text-muted-foreground">{tool.description}</p>
				<p class="mt-5 text-sm font-medium text-primary">Try the tool</p>
			</a>
		{/each}
	</div>
</section>
