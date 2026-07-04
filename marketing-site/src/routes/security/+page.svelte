<script lang="ts">
	import PageHero from '../_components/PageHero.svelte';
	import { securityItems, siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Security - OpenPost</title>
	<meta
		name="description"
		content="OpenPost security details: encrypted provider tokens, passkeys, TOTP, sessions, and revocable automation tokens."
	/>
	<link rel="canonical" href={`${siteUrl}/security`} />
</svelte:head>

<PageHero
	eyebrow="Security"
	title="Security for the accounts that publish on your behalf."
	description="OpenPost is built around practical controls for social tokens, user login, sessions, API tokens, CLI devices, and MCP clients."
	secondaryHref="/open-source"
	secondaryLabel="Open-source trust"
/>

<section class="section-pad">
	<div class="mx-auto grid max-w-7xl gap-4 px-4 sm:px-6 md:grid-cols-2 lg:px-8">
		{#each securityItems as item (item.title)}
			{@const Icon = item.icon}
			<article class="rounded-xl border bg-card p-6">
				<Icon class="size-6 text-primary" />
				<h2 class="mt-5 text-xl font-semibold">{item.title}</h2>
				<p class="mt-3 text-sm leading-7 text-muted-foreground">{item.description}</p>
			</article>
		{/each}
	</div>
</section>
