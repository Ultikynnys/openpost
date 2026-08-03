<!--
THESIS: OpenPost gives a small team one complete content workflow, then makes the publishing cadence visible.
OWN-WORLD: Centered product-led narrative, warm orange proof cells, dark framed product surfaces, and spacious editorial pacing.
STORY: Understand the promise, see the real product, follow the workflow, see public proof, choose a deployment, and start.
FIRST VIEWPORT: A centered outcome statement and two clear routes into a real, working destination composer.
FORM: Product-led publishing workspace with activity squares as the recurring proof and motion language.
-->
<script lang="ts">
  import {
    ArrowRight,
    CalendarRange,
    Check,
    Github,
    Layers3,
    LockKeyhole,
    Server,
  } from "lucide-svelte";
  import { Button } from "$lib/components/ui/button";
  import PlatformIcon from "$lib/components/platform-icon.svelte";
  import DestinationComposerDemo from "./_components/DestinationComposerDemo.svelte";
  import PublishingActivityField from "./_components/PublishingActivityField.svelte";
  import ScrollReveal from "./_components/ScrollReveal.svelte";
  import {
    faqs,
    githubUrl,
    managedAccessSummary,
    managedSignupUrl,
    platforms,
    plans,
    selfHostingDocsUrl,
    siteUrl,
  } from "./_marketing";

  const workflow = [
    {
      number: "01",
      title: "Create the source",
      description:
        "Write a post or edit a video without leaving the workspace.",
    },
    {
      number: "02",
      title: "Adapt each version",
      description:
        "Change copy, media, format, and settings for every destination.",
    },
    {
      number: "03",
      title: "Schedule and follow up",
      description:
        "Publish on time, see each result, and retry only what needs attention.",
    },
  ] as const;

  const productStories = [
    {
      eyebrow: "One workspace",
      title: "The whole publishing workflow stays together.",
      description:
        "Draft posts and threads, manage reusable media, plan the calendar, and check every account from one focused workspace.",
      image: "/assets/screenshots/main-dark.png",
      alt: "OpenPost text-and-thread composer and publication workspace",
      icon: Layers3,
      points: [
        "Text and reply threads",
        "Stories, short video, and video",
        "Calendar and publication status",
      ],
    },
    {
      eyebrow: "Destination controls",
      title: "Each platform gets the version it needs.",
      description:
        "OpenPost keeps the shared idea intact while you adjust the copy, media, limits, and publishing options for each account.",
      image: "/assets/screenshots/accounts-dark.png",
      alt: "OpenPost connected social accounts page",
      icon: LockKeyhole,
      points: [
        "Account-aware limits",
        "Per-destination previews",
        "Encrypted connected-account tokens",
      ],
    },
    {
      eyebrow: "Media studio",
      title: "Prepare the asset where you publish it.",
      description:
        "Keep source files, alt text, favorites, and use history in the media library, then open focused editors for video work.",
      image: "/assets/screenshots/media-dark.png",
      alt: "OpenPost media library with reusable assets",
      icon: CalendarRange,
      points: [
        "Reusable media library",
        "Alt text and metadata",
        "Focused video editing modes",
      ],
    },
  ] as const;

  const featuredPlans = plans.slice(0, 3);
  const shortFaqs = faqs.slice(0, 4);
</script>

<svelte:head>
  <title>OpenPost - Build the company. Keep it visible.</title>
  <meta
    name="description"
    content="Create, adapt, schedule, and track social posts for every account from one open-source workspace. Use OpenPost managed or self-hosted."
  />
  <link rel="canonical" href={siteUrl} />
  <meta name="robots" content="index, follow" />
  <meta property="og:url" content={siteUrl} />
  <meta
    property="og:title"
    content="OpenPost - Build the company. Keep it visible."
  />
  <meta
    property="og:description"
    content="One open-source workspace for the complete social publishing workflow."
  />
  <meta
    name="twitter:title"
    content="OpenPost - Build the company. Keep it visible."
  />
  <meta
    name="twitter:description"
    content="Create, adapt, schedule, and track social posts from one workspace."
  />
</svelte:head>

<section class="hero overflow-hidden">
  <div
    class="marketing-shell pb-16 pt-16 text-center sm:pb-20 sm:pt-24 lg:pt-28"
  >
    <p class="section-label hero-enter hero-enter-1">
      Open-source social publishing
    </p>
    <h1 class="marketing-title hero-enter hero-enter-2 mx-auto mt-5">
      Build the company.<br /><span class="text-primary">Keep it visible.</span>
    </h1>
    <p class="marketing-copy hero-enter hero-enter-3 mx-auto mt-7 max-w-2xl">
      Create, adapt, schedule, and track content for every social account you
      run from one workspace.
    </p>
    <div
      class="hero-enter hero-enter-4 mt-8 flex flex-wrap justify-center gap-3"
    >
      <Button href={managedSignupUrl} size="lg">
        Try OpenPost
        <ArrowRight data-icon="inline-end" />
      </Button>
      <Button href={selfHostingDocsUrl} variant="outline" size="lg">
        <Server data-icon="inline-start" />
        Self-host
      </Button>
    </div>
    <p
      class="hero-enter hero-enter-4 mx-auto mt-4 max-w-xl text-xs leading-5 text-muted-foreground"
    >
      {managedAccessSummary}
    </p>

    <div class="hero-product hero-enter hero-enter-5 mt-12 sm:mt-16">
      <div class="window-bar" aria-hidden="true">
        <span></span><span></span><span></span>
        <p>openpost.social</p>
      </div>
      <DestinationComposerDemo />
    </div>
  </div>
</section>

<section
  aria-label="Supported platforms"
  class="marketing-rule border-y bg-muted/24"
>
  <div class="marketing-shell py-7">
    <p
      class="text-center text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground"
    >
      Publish to the accounts you already run
    </p>
    <div
      class="mt-5 flex items-center gap-x-7 gap-y-3 overflow-x-auto pb-1 sm:flex-wrap sm:justify-center"
    >
      {#each platforms as platform (platform.slug)}
        <a
          href={`/platforms/${platform.slug}`}
          class="focus-ring inline-flex min-h-11 shrink-0 items-center gap-2 rounded-md text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <PlatformIcon platform={platform.short} class="size-4" />
          {platform.name}
        </a>
      {/each}
    </div>
  </div>
</section>

<section class="section-pad" aria-labelledby="workflow-title">
  <div class="marketing-shell">
    <ScrollReveal class="mx-auto max-w-3xl text-center">
      <p class="section-label">From draft to result</p>
      <h2 id="workflow-title" class="marketing-heading mx-auto mt-4">
        Not another scheduler. The complete content workflow.
      </h2>
      <p class="marketing-copy mx-auto mt-6">
        OpenPost keeps the source, every destination version, and the
        publication result connected.
      </p>
    </ScrollReveal>

    <ol class="marketing-rule mt-14 grid border-y md:grid-cols-3">
      {#each workflow as step, index (step.title)}
        <li class="workflow-step py-8 md:px-8 md:first:pl-0 md:last:pr-0">
          <ScrollReveal delay={index * 90}>
            <span class="font-mono text-xs font-semibold text-primary"
              >{step.number}</span
            >
            <h3 class="mt-5 text-xl font-semibold tracking-tight">
              {step.title}
            </h3>
            <p class="mt-3 max-w-sm text-sm leading-6 text-muted-foreground">
              {step.description}
            </p>
          </ScrollReveal>
        </li>
      {/each}
    </ol>
  </div>
</section>

<section
  id="product"
  class="section-pad marketing-rule scroll-mt-24 border-y bg-muted/18"
  aria-labelledby="product-title"
>
  <div class="marketing-shell">
    <ScrollReveal class="max-w-3xl">
      <p class="section-label">Inside OpenPost</p>
      <h2 id="product-title" class="marketing-heading mt-4">
        One system from idea to published post.
      </h2>
    </ScrollReveal>

    <div class="mt-16 grid gap-20 lg:gap-28">
      {#each productStories as story, index (story.title)}
        {@const Icon = story.icon}
        <article
          class="product-story"
          class:product-story-reverse={index % 2 === 1}
        >
          <ScrollReveal class="product-copy" delay={70}>
            <div
              class="grid size-10 place-items-center rounded-xl border bg-background text-primary shadow-sm"
            >
              <Icon class="size-5" aria-hidden="true" />
            </div>
            <p class="section-label mt-6">{story.eyebrow}</p>
            <h3
              class="mt-4 max-w-lg text-3xl font-semibold leading-[1.04] tracking-[-0.035em] text-balance sm:text-4xl"
            >
              {story.title}
            </h3>
            <p class="mt-5 max-w-xl leading-7 text-muted-foreground">
              {story.description}
            </p>
            <ul class="mt-6 grid gap-3 text-sm">
              {#each story.points as point (point)}
                <li class="flex items-center gap-3">
                  <Check class="size-4 text-primary" aria-hidden="true" />
                  {point}
                </li>
              {/each}
            </ul>
          </ScrollReveal>
          <ScrollReveal class="product-shot" delay={index % 2 === 0 ? 140 : 60}>
            <img
              src={story.image}
              alt={story.alt}
              width="1440"
              height="900"
              loading="lazy"
              decoding="async"
            />
          </ScrollReveal>
        </article>
      {/each}
    </div>
  </div>
</section>

<section class="section-pad overflow-hidden" aria-labelledby="activity-title">
  <div class="marketing-shell">
    <div class="grid gap-10 lg:grid-cols-[0.72fr_1.28fr] lg:items-end">
      <ScrollReveal>
        <p class="section-label">Public profiles</p>
        <h2 id="activity-title" class="marketing-heading mt-4">
          Make consistent work visible.
        </h2>
      </ScrollReveal>
      <ScrollReveal delay={100}>
        <p class="marketing-copy lg:ml-auto">
          Choose a username and opt in to a public profile. It shows your post
          activity, current and longest streaks, peak day, and the platforms and
          workspaces you use most.
        </p>
      </ScrollReveal>
    </div>

    <ScrollReveal class="mt-12" delay={90}>
      <PublishingActivityField />
    </ScrollReveal>

    <div class="mt-8 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {#each ["Lifetime posts", "Peak posting day", "Current and longest streak", "Top platforms and workspaces"] as item, index (item)}
        <ScrollReveal delay={index * 70}>
          <div class="profile-proof">
            <i aria-hidden="true"></i>
            <span>{item}</span>
          </div>
        </ScrollReveal>
      {/each}
    </div>
  </div>
</section>

<section
  class="section-pad marketing-rule border-y bg-muted/18"
  aria-labelledby="deployment-title"
>
  <div class="marketing-shell">
    <ScrollReveal class="mx-auto max-w-3xl text-center">
      <p class="section-label">Choose how to run it</p>
      <h2 id="deployment-title" class="marketing-heading mx-auto mt-4">
        Managed for you. Or fully self-hosted.
      </h2>
    </ScrollReveal>

    <div class="mt-12 grid gap-4 lg:grid-cols-2">
      <ScrollReveal>
        <article class="deployment-card bg-card">
          <div>
            <p
              class="font-mono text-xs font-semibold uppercase tracking-wider text-primary"
            >
              Managed
            </p>
            <h3 class="mt-5 text-3xl font-semibold tracking-tight">
              OpenPost Cloud
            </h3>
            <p class="mt-4 max-w-lg leading-7 text-muted-foreground">
              We run the app, storage, updates, and scheduled publishing. Plans
              start at {plans[0].price}/month.
            </p>
          </div>
          <Button href={managedSignupUrl} class="mt-9 w-fit">
            Create an account
            <ArrowRight data-icon="inline-end" />
          </Button>
        </article>
      </ScrollReveal>
      <ScrollReveal delay={90}>
        <article class="deployment-card bg-foreground text-background">
          <div>
            <p
              class="font-mono text-xs font-semibold uppercase tracking-wider text-background/60"
            >
              Open source
            </p>
            <h3 class="mt-5 text-3xl font-semibold tracking-tight">
              Run it yourself
            </h3>
            <p class="mt-4 max-w-lg leading-7 text-background/68">
              Deploy the AGPL server as one binary or container. SQLite and
              local media work by default, and Redis is not required.
            </p>
          </div>
          <div class="mt-9 flex flex-wrap gap-2">
            <Button href={selfHostingDocsUrl} variant="secondary"
              >Self-hosting guide</Button
            >
            <Button
              href={githubUrl}
              variant="ghost"
              class="text-background hover:bg-background/10 hover:text-background"
            >
              <Github data-icon="inline-start" />
              Source
            </Button>
          </div>
        </article>
      </ScrollReveal>
    </div>
  </div>
</section>

<section class="section-pad" aria-labelledby="pricing-title">
  <div class="marketing-shell">
    <ScrollReveal class="mx-auto max-w-3xl text-center">
      <p class="section-label">Managed plans</p>
      <h2 id="pricing-title" class="marketing-heading mx-auto mt-4">
        Start small. Keep the same workflow as you grow.
      </h2>
    </ScrollReveal>

    <div class="mt-12 grid gap-4 lg:grid-cols-3">
      {#each featuredPlans as plan, index (plan.id)}
        <ScrollReveal delay={index * 80}>
          <article class:featured-plan={plan.featured} class="plan-card">
            <div class="flex items-start justify-between gap-4">
              <div>
                <h3 class="text-lg font-semibold">{plan.name}</h3>
                <p class="mt-2 text-sm leading-6 text-muted-foreground">
                  {plan.description}
                </p>
              </div>
              {#if plan.featured}<span class="plan-tag">Popular</span>{/if}
            </div>
            <p class="mt-8 text-4xl font-semibold tracking-[-0.04em]">
              {plan.price}<span
                class="text-sm font-normal tracking-normal text-muted-foreground"
                >/month</span
              >
            </p>
            <ul class="mt-7 grid gap-3 text-sm text-muted-foreground">
              {#each plan.limits.slice(0, 4) as limit (limit)}
                <li class="flex items-center gap-3">
                  <Check
                    class="size-4 text-primary"
                    aria-hidden="true"
                  />{limit}
                </li>
              {/each}
            </ul>
          </article>
        </ScrollReveal>
      {/each}
    </div>
    <div class="mt-7 text-center">
      <Button href="/pricing" variant="outline"
        >Compare all managed plans</Button
      >
    </div>
  </div>
</section>

<section
  class="section-pad marketing-rule border-t bg-muted/18"
  aria-labelledby="faq-title"
>
  <div class="marketing-shell grid gap-12 lg:grid-cols-[0.65fr_1.35fr]">
    <ScrollReveal>
      <p class="section-label">Questions</p>
      <h2 id="faq-title" class="marketing-heading mt-4">Before you start.</h2>
    </ScrollReveal>
    <div class="marketing-rule border-t">
      {#each shortFaqs as item, index (item.question)}
        <ScrollReveal delay={index * 45}>
          <details class="group marketing-rule border-b py-5">
            <summary
              class="focus-ring flex min-h-11 cursor-pointer list-none items-center justify-between gap-4 rounded-md font-medium"
            >
              {item.question}
              <span
                class="text-xl text-muted-foreground transition-transform group-open:rotate-45"
                aria-hidden="true">+</span
              >
            </summary>
            <p
              class="max-w-2xl pb-2 pr-10 text-sm leading-6 text-muted-foreground"
            >
              {item.answer}
            </p>
          </details>
        </ScrollReveal>
      {/each}
    </div>
  </div>
</section>

<section
  class="closing-section overflow-hidden text-center text-white"
  aria-labelledby="closing-title"
>
  <div class="closing-cells" aria-hidden="true"></div>
  <ScrollReveal class="marketing-shell relative py-24 sm:py-32">
    <p
      class="font-mono text-xs font-semibold uppercase tracking-[0.16em] text-primary"
    >
      OpenPost
    </p>
    <h2
      id="closing-title"
      class="mx-auto mt-5 max-w-4xl text-4xl font-semibold leading-[0.98] tracking-[-0.045em] text-balance sm:text-6xl"
    >
      Everything you build deserves an audience.
    </h2>
    <p class="mx-auto mt-6 max-w-xl leading-7 text-white/62">
      Create the source once. Adapt it for each account. Keep publishing without
      adding another disconnected tool.
    </p>
    <div class="mt-8 flex flex-wrap justify-center gap-3">
      <Button href={managedSignupUrl} size="lg">
        Try OpenPost
        <ArrowRight data-icon="inline-end" />
      </Button>
      <Button href={selfHostingDocsUrl} variant="secondary" size="lg"
        >Self-host</Button
      >
    </div>
  </ScrollReveal>
</section>

<style>
  .hero {
    background:
      radial-gradient(
        circle at 50% 26%,
        color-mix(in oklch, var(--primary) 12%, transparent),
        transparent 25rem
      ),
      linear-gradient(
        to bottom,
        transparent,
        color-mix(in oklch, var(--muted) 28%, transparent)
      );
  }

  .hero-enter {
    animation: hero-enter 720ms cubic-bezier(0.16, 1, 0.3, 1) both;
  }

  .hero-enter-1 {
    animation-delay: 40ms;
  }
  .hero-enter-2 {
    animation-delay: 100ms;
  }
  .hero-enter-3 {
    animation-delay: 180ms;
  }
  .hero-enter-4 {
    animation-delay: 250ms;
  }
  .hero-enter-5 {
    animation-delay: 340ms;
  }

  .hero-product {
    overflow: hidden;
    border: 1px solid color-mix(in oklch, var(--foreground) 18%, transparent);
    border-radius: clamp(1rem, 2vw, 1.6rem);
    background: oklch(0.13 0.01 52);
    box-shadow: 0 2.5rem 7rem
      color-mix(in oklch, var(--foreground) 18%, transparent);
    text-align: left;
  }

  .window-bar {
    display: flex;
    align-items: center;
    gap: 0.42rem;
    min-height: 2.7rem;
    padding-inline: 1rem;
    border-bottom: 1px solid rgb(255 255 255 / 0.09);
    color: rgb(255 255 255 / 0.42);
  }

  .window-bar span {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 999px;
    background: rgb(255 255 255 / 0.18);
  }

  .window-bar span:first-child {
    background: oklch(0.67 0.17 43);
  }

  .window-bar p {
    margin-inline: auto;
    padding-right: 2.3rem;
    font-family: ui-monospace, monospace;
    font-size: 0.67rem;
  }

  .workflow-step + .workflow-step {
    border-top: 1px solid var(--border);
  }

  .product-story {
    display: grid;
    gap: 2.5rem;
    align-items: center;
  }

  :global(.product-shot) {
    min-width: 0;
    overflow: hidden;
    padding: clamp(0.55rem, 1.4vw, 0.9rem);
    border: 1px solid color-mix(in oklch, var(--foreground) 16%, transparent);
    border-radius: 1.35rem;
    background: oklch(0.13 0.01 52);
    box-shadow: 0 1.6rem 4.5rem
      color-mix(in oklch, var(--foreground) 12%, transparent);
  }

  :global(.product-shot) img {
    display: block;
    width: 100%;
    aspect-ratio: 16 / 10;
    border-radius: 0.8rem;
    object-fit: contain;
    object-position: top;
  }

  .profile-proof,
  .plan-card,
  .deployment-card {
    border: 1px solid var(--border);
    border-radius: 1rem;
  }

  .profile-proof {
    display: flex;
    min-height: 4.5rem;
    align-items: center;
    gap: 0.8rem;
    padding: 1rem;
    background: var(--card);
    font-size: 0.82rem;
    font-weight: 600;
  }

  .profile-proof i {
    width: 0.78rem;
    height: 0.78rem;
    flex: none;
    border-radius: 0.22rem;
    background: var(--primary);
    box-shadow: 1rem 0 color-mix(in oklch, var(--primary) 58%, var(--muted));
    margin-right: 1rem;
  }

  .deployment-card {
    display: flex;
    min-height: 24rem;
    flex-direction: column;
    justify-content: space-between;
    padding: clamp(1.5rem, 4vw, 3rem);
  }

  .plan-card {
    min-height: 100%;
    padding: 1.5rem;
    background: var(--card);
  }

  .featured-plan {
    border-color: color-mix(in oklch, var(--primary) 60%, var(--border));
    box-shadow: 0 1rem 3.5rem
      color-mix(in oklch, var(--primary) 10%, transparent);
  }

  .plan-tag {
    border-radius: 999px;
    background: color-mix(in oklch, var(--primary) 14%, transparent);
    padding: 0.3rem 0.55rem;
    color: var(--primary);
    font-size: 0.68rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .closing-section {
    position: relative;
    background: oklch(0.13 0.012 50);
  }

  .closing-cells {
    position: absolute;
    inset: 0;
    opacity: 0.25;
    background-image:
      linear-gradient(oklch(0.68 0.16 44 / 0.32) 1px, transparent 1px),
      linear-gradient(90deg, oklch(0.68 0.16 44 / 0.32) 1px, transparent 1px);
    background-size: 2.4rem 2.4rem;
    mask-image: radial-gradient(circle at center, black, transparent 68%);
  }

  @keyframes hero-enter {
    from {
      opacity: 0;
      transform: translateY(1rem);
      filter: blur(5px);
    }
    to {
      opacity: 1;
      transform: none;
      filter: none;
    }
  }

  @media (min-width: 48rem) {
    .workflow-step + .workflow-step {
      border-top: 0;
      border-left: 1px solid var(--border);
    }
  }

  @media (min-width: 64rem) {
    .product-story {
      grid-template-columns: 0.75fr 1.25fr;
      gap: 5rem;
    }
    .product-story-reverse {
      grid-template-columns: 1.25fr 0.75fr;
    }
    .product-story-reverse :global(.product-copy) {
      order: 2;
    }
    .product-story-reverse :global(.product-shot) {
      order: 1;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .hero-enter {
      animation: none;
    }
  }
</style>
