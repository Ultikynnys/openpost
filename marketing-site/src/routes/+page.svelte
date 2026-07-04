<script lang="ts">
	import {
		ArrowRight,
		CalendarClock,
		CheckCircle2,
		ChevronRight,
		Command,
		Database,
		Terminal
	} from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import PlatformIcon from '$lib/components/platform-icon.svelte';
	import TestimonialWall from './_components/TestimonialWall.svelte';
	import {
		appUrl,
		docsUrl,
		faqs,
		githubUrl,
		platforms,
		plans,
		productFeatures,
		securityItems,
		siteUrl,
		testimonials,
		workflowBlocks
	} from './_marketing';

	const heroFlow = [
		{ label: 'Draft', value: 'Base post' },
		{ label: 'Adapt', value: 'Per account' },
		{ label: 'Schedule', value: 'Next slot' },
		{ label: 'Publish', value: 'Visible status' }
	] as const;
</script>

<svelte:head>
	<title>OpenPost - Hosted social publishing for creators and small teams</title>
	<meta
		name="description"
		content="OpenPost Cloud is a focused social publishing workspace for drafting, previewing, scheduling, and automating posts across modern platforms."
	/>
	<link rel="canonical" href={siteUrl} />
	<meta name="robots" content="index, follow" />
	<meta property="og:type" content="website" />
	<meta property="og:url" content={siteUrl} />
	<meta property="og:title" content="OpenPost - Hosted social publishing" />
	<meta
		property="og:description"
		content="Draft once, adapt per platform, schedule with confidence, and keep the queue visible."
	/>
	<meta property="og:image" content={`${siteUrl}/assets/brand/og-image.svg`} />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content="OpenPost - Hosted social publishing" />
	<meta
		name="twitter:description"
		content="Open source social publishing, hosted for creators and small teams."
	/>
</svelte:head>

<section class="relative overflow-hidden border-b">
	<div class="marketing-grid pointer-events-none absolute inset-0"></div>
	<div class="relative mx-auto max-w-7xl px-4 py-16 sm:px-6 sm:py-20 lg:px-8 lg:py-24">
		<div class="grid items-center gap-12 lg:grid-cols-[0.88fr_1.12fr]">
			<div>
				<Badge class="border-primary/25 bg-primary/10 text-primary">
					Source-open publishing, hosted when you want it managed
				</Badge>
				<h1
					class="mt-6 max-w-4xl text-5xl leading-[0.98] font-semibold tracking-tight text-balance sm:text-6xl xl:text-7xl"
				>
					The open social publishing workspace without the enterprise bloat.
				</h1>
				<p class="mt-6 max-w-2xl text-lg leading-8 text-pretty text-muted-foreground sm:text-xl">
					OpenPost Cloud gives creators, indie teams, and small agencies one focused place to
					draft, adapt, preview, schedule, and automate social posts across modern platforms.
				</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href={appUrl} size="lg">
						Start publishing
						<ArrowRight data-icon="inline-end" />
					</Button>
					<Button href={docsUrl} variant="outline" size="lg">
						Self-host instead
						<ChevronRight data-icon="inline-end" />
					</Button>
				</div>
				<div class="mt-7 grid max-w-xl grid-cols-2 gap-3 text-sm sm:grid-cols-4">
					{#each heroFlow as step (step.label)}
						<div class="rounded-lg border bg-card/70 p-3">
							<p class="font-mono text-[0.68rem] tracking-[0.14em] text-muted-foreground uppercase">
								{step.label}
							</p>
							<p class="mt-1 font-medium">{step.value}</p>
						</div>
					{/each}
				</div>
			</div>

			<div class="relative">
				<div class="rounded-xl border bg-card p-2 shadow-[0_28px_90px_oklch(0.08_0.02_45/0.45)]">
					<div class="flex h-9 items-center gap-2 border-b px-3">
						<span class="size-2 rounded-full bg-[oklch(0.67_0.18_28)]"></span>
						<span class="size-2 rounded-full bg-[oklch(0.78_0.15_82)]"></span>
						<span class="size-2 rounded-full bg-[oklch(0.67_0.13_150)]"></span>
						<span class="ml-2 truncate font-mono text-xs text-muted-foreground">
							app.openpost.social
						</span>
					</div>
					<img
						src="/assets/screenshots/main-dark.png"
						alt="OpenPost composer and scheduling dashboard"
						class="mt-2 block w-full rounded-md border"
						width="1280"
						height="832"
						fetchpriority="high"
					/>
				</div>
				<div
					class="absolute right-4 bottom-4 hidden max-w-xs rounded-lg border bg-background/95 p-4 shadow-xl backdrop-blur md:block"
				>
					<p class="font-mono text-xs text-primary">next-slot ready</p>
					<p class="mt-2 text-sm leading-6 text-muted-foreground">
						Queue the approved post into the next workspace slot, with provider warnings checked
						before it lands.
					</p>
				</div>
			</div>
		</div>

		<div class="mt-12 border-y py-5">
			<div class="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
				<p class="text-sm font-medium text-muted-foreground">
					Connect the platforms your audience already uses.
				</p>
				<div class="flex flex-wrap gap-2">
					{#each platforms as platform (platform.slug)}
						<a
							href={`/platforms/${platform.slug}`}
							class="inline-flex items-center gap-2 rounded-full border bg-card px-3 py-2 text-sm font-medium transition hover:bg-muted"
						>
							<PlatformIcon platform={platform.short} class="size-4" />
							<span>{platform.name}</span>
						</a>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[0.78fr_1.22fr]">
			<div>
				<p class="eyebrow">Product workflow</p>
				<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
					Everything important happens before the post is sent.
				</h2>
				<p class="mt-5 max-w-xl text-lg leading-8 text-muted-foreground">
					OpenPost is not trying to become a giant social command center. It focuses on the
					work that makes publishing reliable: writing, adapting, previewing, scheduling, and
					seeing what happened.
				</p>
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				{#each workflowBlocks as block (block.title)}
					{@const Icon = block.icon}
					<article class="rounded-lg border bg-card p-5">
						<Icon class="size-5 text-primary" />
						<h3 class="mt-4 font-semibold">{block.title}</h3>
						<p class="mt-2 text-sm leading-6 text-muted-foreground">{block.description}</p>
					</article>
				{/each}
			</div>
		</div>

		<div class="mt-16 grid gap-6">
			{#each productFeatures as feature, index (feature.title)}
				{@const Icon = feature.icon}
				<article
					class="grid gap-6 rounded-xl border bg-card p-3 lg:grid-cols-[0.78fr_1.22fr] lg:items-center"
				>
					<div class="p-5 lg:p-8 {index % 2 === 1 ? 'lg:order-2' : ''}">
						<p class="eyebrow">{feature.eyebrow}</p>
						<h3 class="mt-4 text-2xl leading-tight font-semibold text-balance sm:text-4xl">
							{feature.title}
						</h3>
						<p class="mt-4 text-base leading-7 text-muted-foreground">{feature.description}</p>
						<div class="mt-6 inline-flex items-center gap-2 rounded-full border px-3 py-2 text-sm">
							<Icon class="size-4 text-primary" />
							<span>Implemented in the app today</span>
						</div>
					</div>
					<figure class="overflow-hidden rounded-lg border bg-background">
						<img
							src={feature.image}
							alt={feature.alt}
							class="block aspect-[16/10] w-full object-cover object-top"
							width="1280"
							height="832"
							loading="lazy"
						/>
					</figure>
				</article>
			{/each}
		</div>
	</div>
</section>

<section class="section-pad border-y bg-muted/25">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="grid gap-10 lg:grid-cols-[1fr_1fr] lg:items-center">
			<div>
				<p class="eyebrow">Automation</p>
				<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
					Web app for editorial work. CLI and MCP for everything around it.
				</h2>
				<p class="mt-5 text-lg leading-8 text-muted-foreground">
					Use the app to review content, then let scripts, CI, cron, or assistants create
					drafts, upload media, choose social sets, and schedule approved posts.
				</p>
				<div class="mt-8 flex flex-wrap gap-3">
					<Button href="/tools" variant="secondary">
						Explore free tools
						<ArrowRight data-icon="inline-end" />
					</Button>
					<Button href={docsUrl} variant="outline">Read automation docs</Button>
				</div>
			</div>
			<div class="grid gap-4">
				<div class="rounded-lg border bg-card p-5 font-mono text-sm leading-7">
					<p class="text-muted-foreground">$ openpost set create launch --accounts x,linkedin</p>
					<p class="text-muted-foreground">$ openpost post create --set launch --schedule next-slot</p>
					<p class="mt-3 text-primary">scheduled for the next workspace slot</p>
				</div>
				<div class="grid gap-4 sm:grid-cols-3">
					{#each [Terminal, Command, Database] as Icon, i (`automation-${i}`)}
						<div class="rounded-lg border bg-card p-4">
							<Icon class="size-5 text-primary" />
							<p class="mt-3 text-sm font-medium">
								{['CLI', 'MCP', 'API'][i]}
							</p>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>

<section class="section-pad">
	<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
		<div class="flex flex-col justify-between gap-6 lg:flex-row lg:items-end">
			<div class="max-w-3xl">
				<p class="eyebrow">Hosted plans</p>
				<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
					Pricing that matches focused publishing, not enterprise procurement.
				</h2>
				<p class="mt-5 text-lg leading-8 text-muted-foreground">
					Every plan includes the same core composer, scheduling workflow, media library, CLI,
					MCP, and source-open trust. Limits scale with workspace size.
				</p>
			</div>
			<Button href="/pricing" variant="outline">
				Full pricing
				<ArrowRight data-icon="inline-end" />
			</Button>
		</div>

		<div class="mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-5">
			{#each plans as plan (plan.id)}
				<article
					class="rounded-lg border bg-card p-5 {plan.featured
						? 'border-primary/50 shadow-[0_18px_50px_oklch(0.12_0.03_45/0.32)]'
						: ''}"
				>
					<div class="flex items-start justify-between gap-4">
						<div>
							<h3 class="font-semibold">{plan.name}</h3>
							<p class="mt-1 text-sm text-muted-foreground">{plan.description}</p>
						</div>
						{#if plan.featured}
							<span class="rounded-full bg-primary px-2 py-1 text-xs text-primary-foreground">
								Popular
							</span>
						{/if}
					</div>
					<p class="mt-5 text-3xl font-semibold">{plan.price}<span class="text-sm text-muted-foreground">/mo</span></p>
					<ul class="mt-5 grid gap-2 text-sm text-muted-foreground">
						{#each plan.limits as limit (limit)}
							<li class="flex gap-2">
								<CheckCircle2 class="mt-0.5 size-4 shrink-0 text-primary" />
								<span>{limit}</span>
							</li>
						{/each}
					</ul>
					<Button href={`${appUrl}/register?plan=${plan.id}`} class="mt-6 w-full" variant={plan.featured ? 'default' : 'outline'}>
						Choose {plan.name}
					</Button>
				</article>
			{/each}
		</div>
	</div>
</section>

<TestimonialWall {testimonials} />

<section class="section-pad">
	<div class="mx-auto grid max-w-7xl gap-10 px-4 sm:px-6 lg:grid-cols-[0.85fr_1.15fr] lg:px-8">
		<div>
			<p class="eyebrow">Trust</p>
			<h2 class="mt-4 text-3xl leading-tight font-semibold text-balance sm:text-5xl">
				Security and openness are part of the product, not a footnote.
			</h2>
			<p class="mt-5 text-lg leading-8 text-muted-foreground">
				OpenPost handles provider tokens and automation credentials with the same practical
				discipline it applies to scheduling.
			</p>
			<Button href="/security" class="mt-8" variant="secondary">Security details</Button>
		</div>
		<div class="grid gap-4 sm:grid-cols-2">
			{#each securityItems as item (item.title)}
				{@const Icon = item.icon}
				<article class="rounded-lg border bg-card p-5">
					<Icon class="size-5 text-primary" />
					<h3 class="mt-4 font-semibold">{item.title}</h3>
					<p class="mt-2 text-sm leading-6 text-muted-foreground">{item.description}</p>
				</article>
			{/each}
		</div>
	</div>
</section>

<section class="border-y bg-primary text-primary-foreground">
	<div class="mx-auto grid max-w-7xl gap-8 px-4 py-12 sm:px-6 lg:grid-cols-[1fr_auto] lg:items-center lg:px-8">
		<div>
			<h2 class="text-3xl font-semibold text-balance">Prefer to run it yourself?</h2>
			<p class="mt-3 max-w-2xl text-base leading-7 text-primary-foreground/75">
				OpenPost is source open, self-hostable, and intentionally lightweight: one binary or
				container, SQLite by default, local media by default, and no Redis queue requirement.
			</p>
		</div>
		<div class="flex flex-wrap gap-3">
			<Button href="/open-source" variant="secondary">Open-source page</Button>
			<Button href={githubUrl} variant="outline" class="border-primary-foreground/25 text-primary-foreground hover:bg-primary-foreground/10">
				View GitHub
			</Button>
		</div>
	</div>
</section>

<section class="section-pad">
	<div class="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
		<p class="eyebrow">FAQ</p>
		<h2 class="mt-4 text-3xl font-semibold text-balance sm:text-5xl">
			The honest answers belong on the page.
		</h2>
		<div class="mt-10 divide-y rounded-xl border bg-card">
			{#each faqs as item (item.question)}
				<details class="group p-5">
					<summary class="flex cursor-pointer list-none items-center justify-between gap-4 font-semibold">
						<span>{item.question}</span>
						<CalendarClock class="size-4 text-muted-foreground transition group-open:rotate-45" />
					</summary>
					<p class="mt-4 max-w-3xl text-sm leading-7 text-muted-foreground">{item.answer}</p>
				</details>
			{/each}
		</div>
	</div>
</section>

<section class="section-pad border-t">
	<div class="mx-auto max-w-5xl px-4 text-center sm:px-6 lg:px-8">
		<p class="eyebrow">Start focused</p>
		<h2 class="mt-4 text-4xl leading-tight font-semibold text-balance sm:text-6xl">
			Queue the next post without adopting a social media suite.
		</h2>
		<p class="mx-auto mt-5 max-w-2xl text-lg leading-8 text-muted-foreground">
			Open the hosted app, create the first workspace, connect the platforms that matter, and
			keep publishing visible.
		</p>
		<div class="mt-8 flex justify-center gap-3">
			<Button href={appUrl} size="lg">
				Open app
				<ArrowRight data-icon="inline-end" />
			</Button>
			<Button href="/pricing" variant="outline" size="lg">See plans</Button>
		</div>
	</div>
</section>
