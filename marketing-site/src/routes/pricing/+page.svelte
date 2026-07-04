<script lang="ts">
	import PageHero from '../_components/PageHero.svelte';
	import PricingComparison from '../_components/PricingComparison.svelte';
	import { siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Pricing - OpenPost</title>
	<meta
		name="description"
		content="OpenPost pricing for creators, small teams, and agencies."
	/>
	<link rel="canonical" href={`${siteUrl}/pricing`} />
</svelte:head>

<PageHero
	eyebrow="Pricing"
	title="Plans for focused publishing teams."
	description="Start with one workspace and scale to agencies without paying for a social media suite."
	secondaryHref="/platforms"
	secondaryLabel="See platforms"
/>

<PricingComparison />
