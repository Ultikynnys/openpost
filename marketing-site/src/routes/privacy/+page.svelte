<script lang="ts">
	import PageHero from '../_components/PageHero.svelte';
	import { siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Privacy - OpenPost</title>
	<meta name="description" content="OpenPost privacy summary." />
	<link rel="canonical" href={`${siteUrl}/privacy`} />
</svelte:head>

<PageHero
	eyebrow="Legal"
	title="Privacy summary."
	description="A practical placeholder for the hosted privacy policy while the legal text is finalized."
	secondaryHref="/security"
	secondaryLabel="Security details"
/>

<section class="section-pad">
	<div class="prose prose-invert mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
		<p>
			OpenPost Cloud stores the account, workspace, media, scheduling, billing, and automation data
			needed to provide the hosted publishing service.
		</p>
		<p>
			Provider tokens are encrypted at rest. Users can manage sessions, passkeys, TOTP, and
			revocable CLI/API/MCP tokens from account settings.
		</p>
		<p>
			This page should be replaced with final legal language before broad public launch.
		</p>
	</div>
</section>
