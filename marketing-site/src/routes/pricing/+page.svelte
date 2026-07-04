<script lang="ts">
	import { CheckCircle2 } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import PageHero from '../_components/PageHero.svelte';
	import { appUrl, plans, siteUrl } from '../_marketing';
</script>

<svelte:head>
	<title>Pricing - OpenPost</title>
	<meta
		name="description"
		content="OpenPost Cloud pricing for creators, small teams, and agencies."
	/>
	<link rel="canonical" href={`${siteUrl}/pricing`} />
</svelte:head>

<PageHero
	eyebrow="Pricing"
	title="Hosted plans for focused publishing teams."
	description="Start with one workspace and scale to agencies without paying for enterprise features OpenPost does not pretend to be."
	secondaryHref="/platforms"
	secondaryLabel="See platforms"
/>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
			{#each plans as plan (plan.id)}
				<article class="rounded-xl border bg-card p-5 {plan.featured ? 'border-primary/60' : ''}">
					<div class="min-h-28">
						<div class="flex items-center justify-between gap-3">
							<h2 class="text-xl font-semibold">{plan.name}</h2>
							{#if plan.featured}
								<span class="rounded-full bg-primary px-2 py-1 text-xs text-primary-foreground">
									Popular
								</span>
							{/if}
						</div>
						<p class="mt-3 text-sm leading-6 text-muted-foreground">{plan.description}</p>
					</div>
					<p class="mt-6 text-4xl font-semibold">
						{plan.price}<span class="text-base text-muted-foreground">/mo</span>
					</p>
					<ul class="mt-6 grid gap-3 text-sm text-muted-foreground">
						{#each plan.limits as limit (limit)}
							<li class="flex gap-2">
								<CheckCircle2 class="mt-0.5 size-4 shrink-0 text-primary" />
								<span>{limit}</span>
							</li>
						{/each}
					</ul>
					<Button
						href={`${appUrl}/register?plan=${plan.id}`}
						class="mt-8 w-full"
						variant={plan.featured ? 'default' : 'outline'}
					>
						Start {plan.name}
					</Button>
				</article>
			{/each}
		</div>

		<div class="mt-12 grid gap-4 rounded-xl border bg-muted/25 p-5 md:grid-cols-3">
			<div>
				<h2 class="font-semibold">What every plan includes</h2>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					Composer, platform variants, media library, scheduling, social sets, CLI, MCP, and
					security controls.
				</p>
			</div>
			<div>
				<h2 class="font-semibold">What is intentionally absent</h2>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					No analytics promise, no enterprise approval suite, and no universal video parity
					claim.
				</p>
			</div>
			<div>
				<h2 class="font-semibold">Need to self-host?</h2>
				<p class="mt-2 text-sm leading-6 text-muted-foreground">
					The source-open project remains lightweight to run yourself. Cloud is the managed path.
				</p>
			</div>
		</div>

		<div class="mt-10 flex justify-center">
			<Button href={appUrl} size="lg">Open hosted app</Button>
		</div>
	</div>
</section>
