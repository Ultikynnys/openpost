<script module lang="ts">
	const appUrl = 'https://app.openpost.social';
	const docsUrl = 'https://docs.openpost.social';
	const githubUrl = 'https://github.com/rodrgds/openpost';

	const destinations = [
		{
			label: 'Open app',
			title: 'OpenPost app',
			description: 'Go to the hosted OpenPost workspace and start publishing.',
			href: appUrl,
			meta: 'app.openpost.social'
		},
		{
			label: 'Read docs',
			title: 'Documentation',
			description: 'Install, configure, self-host, and integrate OpenPost.',
			href: docsUrl,
			meta: 'docs.openpost.social'
		}
	];
</script>

<script lang="ts">
	import { ArrowUpRight, BookOpenText, Github, LogIn } from 'lucide-svelte';
	import Logo from '$lib/components/Logo.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
</script>

<svelte:head>
	<title>OpenPost - Open source social publishing</title>
	<meta
		name="description"
		content="OpenPost is an open source social publishing tool. Open the hosted app or read the documentation for setup, self-hosting, and integrations."
	/>
	<link rel="canonical" href="https://openpost.social/" />
	<meta name="robots" content="index, follow" />
	<meta property="og:type" content="website" />
	<meta property="og:url" content="https://openpost.social/" />
	<meta property="og:title" content="OpenPost - Open source social publishing" />
	<meta
		property="og:description"
		content="Open the hosted OpenPost app or read the documentation for setup, self-hosting, and integrations."
	/>
	<meta property="og:image" content="https://openpost.social/logo.svg" />
	<meta name="twitter:card" content="summary" />
	<meta name="twitter:title" content="OpenPost - Open source social publishing" />
	<meta
		name="twitter:description"
		content="Open the hosted OpenPost app or read the documentation for setup, self-hosting, and integrations."
	/>
</svelte:head>

<main class="min-h-dvh bg-background text-foreground">
	<div class="mx-auto flex min-h-dvh w-full max-w-5xl flex-col px-4 py-5 sm:px-6 lg:px-8">
		<header class="flex items-center justify-between gap-4 border-b border-border pb-4">
			<a class="inline-flex items-center gap-2" href="/" aria-label="OpenPost home">
				<Logo width={36} height={28} />
				<span class="text-sm font-semibold tracking-tight">OpenPost</span>
			</a>
			<nav class="flex items-center gap-1" aria-label="Primary">
				<Button variant="ghost" size="sm" href={docsUrl}>Docs</Button>
				<Button variant="ghost" size="sm" href={githubUrl} target="_blank" rel="noreferrer">
					<Github aria-hidden="true" />
					Source
				</Button>
			</nav>
		</header>

		<section class="grid flex-1 items-center gap-10 py-14 md:grid-cols-[minmax(0,1fr)_22rem] md:py-20">
			<div class="max-w-2xl">
				<Badge class="mb-5 bg-primary/10 text-primary ring-1 ring-primary/15">
					Open source social publishing
				</Badge>
				<h1 class="max-w-xl text-4xl leading-tight font-semibold tracking-tight sm:text-5xl">
					OpenPost
				</h1>
				<p class="mt-4 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">
					A focused place to draft, schedule, and publish social posts. This site is now a
					thin front door while the product and docs carry the real work.
				</p>

				<div class="mt-7 flex flex-wrap gap-3">
					<Button size="lg" href={appUrl}>
						<LogIn aria-hidden="true" />
						Open app
					</Button>
					<Button variant="outline" size="lg" href={docsUrl}>
						<BookOpenText aria-hidden="true" />
						Read docs
					</Button>
				</div>
			</div>

			<div class="grid gap-3" aria-label="OpenPost destinations">
				{#each destinations as destination (destination.href)}
					<Card class="bg-card/80">
						<CardHeader>
							<div class="flex items-start justify-between gap-4">
								<div>
									<CardTitle>{destination.title}</CardTitle>
									<CardDescription class="mt-1">{destination.meta}</CardDescription>
								</div>
								<ArrowUpRight class="size-4 text-muted-foreground" aria-hidden="true" />
							</div>
						</CardHeader>
						<CardContent>
							<p class="text-sm leading-6 text-muted-foreground">{destination.description}</p>
							<Button
								class="mt-4 w-full justify-between"
								variant="secondary"
								href={destination.href}
							>
								{destination.label}
								<ArrowUpRight aria-hidden="true" />
							</Button>
						</CardContent>
					</Card>
				{/each}
			</div>
		</section>
	</div>
</main>
