<script lang="ts">
	import { ArrowRight, Minus, CheckCircle2 } from 'lucide-svelte';
	import PageHero from '../_components/PageHero.svelte';
	import { comparisons, siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Compare OpenPost - OpenPost</title>
	<meta
		name="description"
		content="Compare OpenPost with Buffer, Hootsuite, Typefully, Postiz, Post Bridge, and Mixpost."
	/>
	<link rel="canonical" href={`${siteUrl}/compare`} />
</svelte:head>

<PageHero
	eyebrow="Compare"
	title="A comparison hub that does not hide the tradeoffs."
	description="OpenPost is not the biggest social suite. It is the focused, source-open publishing workspace for people who care about queue visibility, automation, and lightweight operation."
	secondaryHref="/pricing"
	secondaryLabel="See pricing"
/>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each comparisons as comparison (comparison.slug)}
				<a href={`/compare/${comparison.slug}`} class="group rounded-xl border bg-card p-5 transition hover:bg-muted/30">
					<div class="flex items-center justify-between gap-4">
						<h2 class="font-semibold">OpenPost vs {comparison.name}</h2>
						<ArrowRight class="size-4 text-muted-foreground transition group-hover:translate-x-0.5 group-hover:text-foreground" />
					</div>
					<p class="mt-4 text-sm leading-6 text-muted-foreground">{comparison.openPostAngle}</p>
				</a>
			{/each}
		</div>

		<div class="mt-12 overflow-hidden rounded-xl border bg-card">
			<div class="grid grid-cols-[1.2fr_0.8fr_0.8fr] border-b p-4 text-sm font-semibold">
				<span>Capability</span>
				<span>OpenPost</span>
				<span>Be careful with claims</span>
			</div>
			{#each [
				['Core scheduling', 'Included', 'Yes'],
				['Open-source codebase', 'Yes', 'Not universal'],
				['Analytics', 'Not a launch feature', 'Often over-claimed'],
				['Video everywhere', 'Provider-dependent', 'Usually nuanced'],
				['Enterprise approvals', 'Not the focus', 'Suite territory']
			] as row (row[0])}
				<div class="grid grid-cols-[1.2fr_0.8fr_0.8fr] gap-3 border-b p-4 text-sm last:border-b-0">
					<span>{row[0]}</span>
					<span class="flex items-center gap-2 text-muted-foreground">
						{#if row[1] === 'Yes' || row[1] === 'Included'}
							<CheckCircle2 class="size-4 text-primary" />
						{:else}
							<Minus class="size-4" />
						{/if}
						{row[1]}
					</span>
					<span class="text-muted-foreground">{row[2]}</span>
				</div>
			{/each}
		</div>
	</div>
</section>
