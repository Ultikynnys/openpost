<script lang="ts">
	import PageHero from '../_components/PageHero.svelte';
	import { siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Terms - OpenPost</title>
	<meta name="description" content="OpenPost terms summary." />
	<link rel="canonical" href={`${siteUrl}/terms`} />
</svelte:head>

<PageHero
	eyebrow="Legal"
	title="Terms summary."
	description="A practical placeholder for hosted terms while the final legal text is prepared."
	secondaryHref="/pricing"
	secondaryLabel="Pricing"
/>

<section class="section-pad">
	<div class="prose prose-invert mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
		<p>
			OpenPost Cloud is a hosted publishing service for drafting, scheduling, and automating social
			posts through connected provider accounts.
		</p>
		<p>
			Provider API limits, app review requirements, rate limits, and media rules vary by platform.
			OpenPost does not promise universal provider feature parity.
		</p>
		<p>
			This page should be replaced with final legal language before broad public launch.
		</p>
	</div>
</section>
