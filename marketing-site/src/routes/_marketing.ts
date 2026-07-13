import {
	Activity,
	CalendarClock,
	CheckCircle2,
	Clock3,
	Code2,
	Database,
	FileText,
	GitBranch,
	KeyRound,
	Library,
	LockKeyhole,
	MessageSquareText,
	PanelTop,
	ShieldCheck,
	Terminal,
	UsersRound,
	Workflow
} from 'lucide-svelte';
import { PLATFORM_LIMITS, publicPlatformLimits } from '$lib/platform-limits';

export const appUrl = 'https://app.openpost.social';
export const userDocsUrl = 'https://docs.openpost.social/usage/';
export const selfHostingDocsUrl = 'https://docs.openpost.social/self-hosting/';
export const developerDocsUrl = 'https://docs.openpost.social/development/';
export const docsUrl = userDocsUrl;
export const githubUrl = 'https://github.com/rodrgds/openpost';
export const siteUrl = 'https://openpost.social';

export const navItems = [
	{ label: 'Pricing', href: '/#pricing' },
	{ label: 'Platforms', href: '/platforms' },
	{ label: 'Compare', href: '/compare' },
	{ label: 'Tools', href: '/tools' },
	{ label: 'Security', href: '/security' },
	{ label: 'Docs', href: docsUrl }
] as const;

export const planIDs = ['starter', 'creator', 'pro', 'team', 'agency'] as const;

export const plans = [
	{
		id: 'starter',
		name: 'Starter',
		price: '€6',
		description: 'Small projects that need managed posting without extra workspace overhead.',
		limits: ['1 workspace', '3 social accounts', '100 scheduled posts/month', '1 GB media'],
		featured: false
	},
	{
		id: 'creator',
		name: 'Creator',
		price: '€12',
		description: 'For active creators and operator-led brands publishing every week.',
		limits: ['3 workspaces', '6 social accounts', '500 scheduled posts/month', '5 GB media'],
		featured: true
	},
	{
		id: 'pro',
		name: 'Pro',
		price: '€24',
		description: 'Higher limits for teams, heavier media use, and larger publishing operations.',
		limits: ['10 workspaces', '15 social accounts', '2,500 scheduled posts/month', '25 GB media'],
		featured: false
	},
	{
		id: 'team',
		name: 'Team',
		price: '€49',
		description: 'Seat-based collaboration for small teams and multi-brand operators.',
		limits: [
			'10 workspaces',
			'25 social accounts',
			'5,000 scheduled posts/month',
			'3 included seats'
		],
		featured: false
	},
	{
		id: 'agency',
		name: 'Agency',
		price: '€99',
		description: 'Agency workspace management with higher account and media limits.',
		limits: [
			'50 workspaces',
			'150 social accounts',
			'25,000 scheduled posts/month',
			'5 included seats'
		],
		featured: false
	}
] as const;

export const platforms = [
	{
		slug: 'x',
		name: 'X',
		short: 'x',
		tag: 'Text, images, threads',
		status: 'Core',
		description:
			'Write short posts, image posts, and reply-style threads for connected X accounts.',
		limits: [
			`${PLATFORM_LIMITS.x.charLimit} characters by default`,
			'25,000 characters for X Premium longer posts',
			PLATFORM_LIMITS.x.media,
			'API/scheduling support for long posts should be verified per account'
		]
	},
	{
		slug: 'mastodon',
		name: 'Mastodon',
		short: 'mastodon',
		tag: 'Fediverse scheduling',
		status: 'Core',
		description:
			'Connect configured or custom public Mastodon instances and schedule posts into your workspace queue.',
		limits: [
			`${PLATFORM_LIMITS.mastodon.charLimit} characters by default`,
			PLATFORM_LIMITS.mastodon.media,
			'Reply chains',
			'Instance-specific behavior'
		]
	},
	{
		slug: 'bluesky',
		name: 'Bluesky',
		short: 'bluesky',
		tag: 'App-password login',
		status: 'Core',
		description:
			'Connect with a Bluesky handle and app password, then publish posts and reply chains.',
		limits: [
			`${PLATFORM_LIMITS.bluesky.charLimit} characters`,
			PLATFORM_LIMITS.bluesky.media,
			'AT Protocol replies',
			'MP4 video path is partial'
		]
	},
	{
		slug: 'linkedin',
		name: 'LinkedIn',
		short: 'linkedin',
		tag: 'Professional posts',
		status: 'Core',
		description:
			'Publish professional updates and account-specific variants without leaving the shared composer.',
		limits: [
			`${PLATFORM_LIMITS.linkedin.charLimit} characters`,
			PLATFORM_LIMITS.linkedin.media,
			'Thread children as comments',
			'App review may apply'
		]
	},
	{
		slug: 'threads',
		name: 'Threads',
		short: 'threads',
		tag: 'Public media required',
		status: 'Core',
		description:
			'Publish Threads posts and reply chains, with media served through public hosted URLs.',
		limits: [
			`${PLATFORM_LIMITS.threads.charLimit} characters`,
			PLATFORM_LIMITS.threads.media,
			'Reply chains',
			'Public media URL required'
		]
	},
	{
		slug: 'facebook',
		name: 'Facebook Pages',
		short: 'facebook',
		tag: 'Pages first slice',
		status: 'First slice',
		description:
			'Connect Facebook Pages through the provider app registry and schedule text or one media attachment.',
		limits: [
			`${PLATFORM_LIMITS.facebook.charLimit.toLocaleString()} characters`,
			PLATFORM_LIMITS.facebook.media,
			'No thread support',
			'Live-account verification matters'
		]
	},
	{
		slug: 'instagram',
		name: 'Instagram Business',
		short: 'instagram',
		tag: 'Media-first',
		status: 'First slice',
		description:
			'Schedule one image or Reel-style video to Instagram Business accounts behind Facebook Pages.',
		limits: [
			`${PLATFORM_LIMITS.instagram.charLimit.toLocaleString()} caption characters`,
			PLATFORM_LIMITS.instagram.media,
			'Business accounts only',
			'No text-only posts'
		]
	},
	{
		slug: 'tiktok',
		name: 'TikTok',
		short: 'tiktok',
		tag: 'Video-first',
		status: 'First slice',
		description:
			'Schedule video-first TikTok posts through a public media URL and provider app configuration.',
		limits: [
			`${PLATFORM_LIMITS.tiktok.charLimit.toLocaleString()} caption characters`,
			PLATFORM_LIMITS.tiktok.media,
			'No image posts',
			'Provider review may apply'
		]
	},
	{
		slug: 'youtube',
		name: 'YouTube',
		short: 'youtube',
		tag: 'Private video upload',
		status: 'First slice',
		description:
			'Upload one scheduled video to a connected channel, private by default in the current adapter path.',
		limits: [
			`${PLATFORM_LIMITS.youtube.charLimit.toLocaleString()} description characters`,
			PLATFORM_LIMITS.youtube.media,
			'Private upload path',
			'YouTube Data API required'
		]
	}
] as const;

export const platformLimitSummaries = publicPlatformLimits();

export const productFeatures = [
	{
		eyebrow: 'Composer',
		title: 'One base post, platform-specific variants when the copy needs to split.',
		description:
			'Draft the canonical message once, then adjust copy and media per account without losing the shared source.',
		icon: MessageSquareText,
		image: '/assets/screenshots/main-dark.png',
		alt: 'OpenPost composer and schedule calendar'
	},
	{
		eyebrow: 'Preview',
		title: 'See the destination shape before the queue sees it.',
		description:
			'Previews render per selected account, and media warnings surface provider limits before scheduling.',
		icon: PanelTop,
		image: '/assets/screenshots/accounts-dark.png',
		alt: 'OpenPost accounts and provider cards'
	},
	{
		eyebrow: 'Media',
		title: 'A reusable media library for scheduled work, not a throwaway upload field.',
		description:
			'Upload once, track usage, keep alt text, favorite useful assets, and avoid deleting media still tied to scheduled posts.',
		icon: Library,
		image: '/assets/screenshots/media-dark.png',
		alt: 'OpenPost media library'
	},
	{
		eyebrow: 'Operations',
		title: 'Activity and failure states are visible instead of hidden.',
		description:
			'OpenPost tracks drafts, scheduled posts, publishing jobs, published posts, and failures so operators can see what needs attention.',
		icon: Activity,
		image: '/assets/screenshots/settings-dark.png',
		alt: 'OpenPost settings and scheduling controls'
	}
] as const;

export const workflowBlocks = [
	{
		title: 'Workspaces',
		description: 'Separate brands, clients, products, media, prompts, schedules, accounts, and team access.',
		icon: UsersRound
	},
	{
		title: 'Posting slots',
		description: 'Use workspace timezones, next-slot scheduling, week starts, draft gaps, and natural delays.',
		icon: CalendarClock
	},
	{
		title: 'Social sets',
		description: 'Save reusable account groups and make one the default for fast posting and automation.',
		icon: Workflow
	},
	{
		title: 'CLI and MCP',
		description: 'Create posts from scripts, cron, CI, or assistants using revocable API and MCP tokens.',
		icon: Terminal
	}
] as const;

export const securityItems = [
	{
		title: 'Encrypted provider tokens',
		description: 'Social access and refresh tokens are encrypted at rest with AES-256-GCM.',
		icon: LockKeyhole
	},
	{
		title: 'Account hardening',
		description: 'Users can enable TOTP, passkeys, and manage active browser sessions.',
		icon: ShieldCheck
	},
	{
		title: 'Revocable automation',
		description: 'CLI, CI, cron, and MCP clients use dedicated tokens with workspace-aware boundaries.',
		icon: KeyRound
	},
	{
		title: 'Open implementation',
		description: 'The API, queue, billing, storage, and provider adapter code are inspectable in the repo.',
		icon: Code2
	}
] as const;

export const tools = [
	{
		slug: 'multi-platform-character-counter',
		name: 'Multi-platform character counter',
		description:
			'Paste once and see character limits for X, Mastodon, Bluesky, LinkedIn, Threads, and Facebook.',
		icon: FileText
	},
	{
		slug: 'post-preview-generator',
		name: 'Post preview generator',
		description:
			'Preview how a post and media could render across selected destinations before you schedule.',
		icon: PanelTop
	},
	{
		slug: 'thread-splitter',
		name: 'Thread splitter',
		description:
			'Turn long copy into platform-aware thread slices that are easier to review and schedule.',
		icon: GitBranch
	},
	{
		slug: 'fediverse-handle-checker',
		name: 'Fediverse handle checker',
		description:
			'Check Mastodon-style and Bluesky-style handles before adding them to launch plans.',
		icon: CheckCircle2
	},
	{
		slug: 'linkedin-text-formatter',
		name: 'LinkedIn text formatter',
		description:
			'Prepare readable LinkedIn copy with lightweight formatting and length awareness.',
		icon: MessageSquareText
	},
	{
		slug: 'best-time-to-post-calculator',
		name: 'Best time to post calculator',
		description:
			'Translate timezone and cadence preferences into reusable posting slots.',
		icon: Clock3
	}
] as const;

export const comparisons = [
	{
		slug: 'buffer',
		name: 'Buffer',
		bestFor: 'Teams that want a mature hosted suite with analytics.',
		openPostAngle: 'OpenPost focuses on a cleaner composer, source-open trust, automation, and lower app plans.',
		pricing: 'Buffer commonly scales by channel and team needs; OpenPost plans scale by workspaces, accounts, scheduled posts, and media.',
		chooseOpenPost: [
			'You want a focused scheduler without a broad engagement suite.',
			'You care about source-open implementation and self-hosting as a fallback.',
			'You want CLI, MCP, and API workflows beside the web app.'
		],
		chooseThem: [
			'You need mature analytics and engagement reporting today.',
			'You want a long-established hosted social suite with a larger ecosystem.'
		]
	},
	{
		slug: 'hootsuite',
		name: 'Hootsuite',
		bestFor: 'Larger teams that need a broad enterprise social management suite.',
		openPostAngle: 'OpenPost is intentionally lighter: scheduling, workspaces, media, CLI/MCP, and fewer enterprise layers.',
		pricing: 'Hootsuite is built for larger social teams and procurement. OpenPost keeps plan choices smaller and easier to reason about.',
		chooseOpenPost: [
			'You publish across a few brands, clients, or products and want less process.',
			'You want queue visibility, workspaces, and automation without enterprise overhead.',
			'You prefer a small app your team can understand quickly.'
		],
		chooseThem: [
			'You need advanced approvals, governance, listening, and reporting.',
			'Your organization already standardizes on enterprise social management.'
		]
	},
	{
		slug: 'typefully',
		name: 'Typefully',
		bestFor: 'Creators focused on X-style writing and threads.',
		openPostAngle: 'OpenPost broadens the workflow across workspaces, provider variants, media, CLI, MCP, and open-source deployment.',
		pricing: 'Typefully is strongest for writing-first creator workflows. OpenPost focuses on multi-platform publishing operations.',
		chooseOpenPost: [
			'You need workspaces, social sets, media reuse, and provider-specific previews.',
			'You publish beyond X-style threads.',
			'You want app, CLI, API, and MCP entry points.'
		],
		chooseThem: [
			'Your main workflow is writing and refining X/LinkedIn-style posts.',
			'You want a polished creator-writing product more than an operations surface.'
		]
	},
	{
		slug: 'postiz',
		name: 'Postiz',
		bestFor: 'Teams looking for an AI-heavy open source social suite.',
		openPostAngle: 'OpenPost leads with a focused publishing workflow, visible queue behavior, and a small operational footprint.',
		pricing: 'Both projects have open-source roots. OpenPost positions the managed app as the primary path while staying smaller in product scope.',
		chooseOpenPost: [
			'You want a quieter publishing app with fewer AI-first surfaces.',
			'You care about Go/SvelteKit, a compact service, and no Redis queue requirement.',
			'You want honest provider caveats shown before scheduling.'
		],
		chooseThem: [
			'You want an AI-heavy social scheduler with broader growth tooling.',
			'You prefer Postiz’s wider all-in-one product direction.'
		]
	},
	{
		slug: 'post-bridge',
		name: 'Post Bridge',
		bestFor: 'Users who want an all-in-one hosted scheduler with many growth tools.',
		openPostAngle: 'OpenPost keeps the core scheduler honest, technical, and source-open, with self-hosting as a trust signal.',
		pricing: 'Post Bridge leans into hosted growth utilities. OpenPost leans into scheduling, workspaces, automation, and source-open trust.',
		chooseOpenPost: [
			'You want to inspect or self-host the implementation if needed.',
			'You want CLI/MCP workflows for technical operators.',
			'You prefer focused publishing over a broad tool bundle.'
		],
		chooseThem: [
			'You want a larger hosted marketing-tool bundle.',
			'You value growth utilities more than source-open operation.'
		]
	},
	{
		slug: 'mixpost',
		name: 'Mixpost',
		bestFor: 'Self-hosted operators who want a larger PHP-based social publishing app.',
		openPostAngle: 'OpenPost is Go and SvelteKit, one binary/container, no Redis queue, and built around a compact cloud path.',
		pricing: 'Mixpost is a strong self-hosted comparison. OpenPost emphasizes the managed app first, while keeping the self-hosting path lightweight.',
		chooseOpenPost: [
			'You want the managed app as the primary path.',
			'You prefer Go/SvelteKit and a single binary/container deployment model.',
			'You want queue, media, and automation primitives without operating Redis.'
		],
		chooseThem: [
			'You specifically want Mixpost’s PHP/Laravel ecosystem.',
			'You want a larger self-hosted social-management app.'
		]
	}
] as const;

export const testimonials = [
	{
		id: 'mara',
		name: 'Mara Lopes',
		role: 'Indie SaaS founder',
		avatar:
			'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=160&h=160&q=80',
		content:
			'OpenPost feels like the scheduler I wanted after outgrowing manual posting. The workspace split keeps product launches and personal channels from bleeding into each other.',
		source: 'X'
	},
	{
		id: 'eli',
		name: 'Eli Mercer',
		role: 'Developer advocate',
		avatar:
			'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=160&h=160&q=80',
		content:
			'I like that it does not pretend every provider works the same. The preview and warning model is exactly what I want before scheduling a week of posts.',
		source: 'LinkedIn'
	},
	{
		id: 'nina',
		name: 'Nina Costa',
		role: 'Content operator',
		avatar:
			'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=160&h=160&q=80',
		content:
			'Social sets and next-slot scheduling are the details that make it feel fast. I can queue the same update across a few accounts without rebuilding the destination list every time.',
		source: 'Mastodon'
	},
	{
		id: 'sam',
		name: 'Samir Patel',
		role: 'Agency operator',
		avatar:
			'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=160&h=160&q=80',
		content:
			'The product is small in the right way. Workspaces, media, schedules, and team invites are there, but it does not feel like an enterprise tool got dropped on a tiny team.',
		source: 'Bluesky'
	},
	{
		id: 'jules',
		name: 'Jules Armand',
		role: 'Open-source maintainer',
		avatar:
			'https://images.unsplash.com/photo-1560250097-0b93528c311a?auto=format&fit=crop&w=160&h=160&q=80',
		content:
			'The CLI and MCP angle is the reason I keep coming back. Publishing from release scripts and letting an assistant prepare drafts against real account boundaries is useful.',
		source: 'GitHub'
	},
	{
		id: 'lena',
		name: 'Lena Wright',
		role: 'Creator',
		avatar:
			'https://images.unsplash.com/photo-1487412720507-e7ab37603c6f?auto=format&fit=crop&w=160&h=160&q=80',
		content:
			'It is refreshing to see a scheduler that is direct about analytics and video limits instead of selling a wall of checkmarks. The core writing flow is what I needed first.',
		source: 'X'
	}
] as const;

export const faqs = [
	{
		question: 'Is using OpenPost different from self-hosting?',
		answer:
			'The app at app.openpost.social is the managed version. The project is still source-open and self-hostable, but the landing page focuses on the ready-to-use workflow.'
	},
	{
		question: 'Does OpenPost include analytics?',
		answer:
			'Advanced engagement analytics are not a launch feature. OpenPost focuses on drafting, previewing, scheduling, queue visibility, media, workspaces, CLI, and MCP first.'
	},
	{
		question: 'Does video publishing work everywhere?',
		answer:
			'Video is provider-dependent. Some adapter paths exist, but each provider has different rules, review requirements, media limits, and public URL needs.'
	},
	{
		question: 'Can I bring my own provider app credentials?',
		answer:
			'Yes. Operators can configure provider apps, and self-hosted users can use environment configuration or the provider app registry depending on the provider.'
	},
	{
		question: 'What happens if a post fails?',
		answer:
			'Failures are kept visible through post and job state. OpenPost is designed for operators who want to see what happened instead of guessing.'
	}
] as const;

export const changelogEntries = [
	{
		date: '2026-07',
		title: 'Plan and billing foundation',
		detail:
			'Plan IDs, Polar checkout, organization billing status, usage counters, and customer portal flows are aligned around Starter, Creator, Pro, Team, and Agency.'
	},
	{
		date: '2026-07',
		title: 'Provider expansion',
		detail:
			'Provider docs and adapters now cover Facebook Pages, Instagram Business, TikTok, and YouTube alongside X, Mastodon, Bluesky, LinkedIn, and Threads.'
	},
	{
		date: '2026-06',
		title: 'CLI, MCP, and next-slot scheduling',
		detail:
			'The CLI and MCP surfaces support social sets, workspace-bound tokens, assistant scheduling tools, and next available posting slots.'
	}
] as const;

export function getPlatform(slug: string) {
	return platforms.find((platform) => platform.slug === slug);
}

export function getComparison(slug: string) {
	return comparisons.find((comparison) => comparison.slug === slug);
}

export function getTool(slug: string) {
	return tools.find((tool) => tool.slug === slug);
}

export type PlatformSlug = (typeof platforms)[number]['slug'];
export type ComparisonSlug = (typeof comparisons)[number]['slug'];
export type ToolSlug = (typeof tools)[number]['slug'];
