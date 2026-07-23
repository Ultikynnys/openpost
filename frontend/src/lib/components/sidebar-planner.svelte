<script lang="ts">
	import { getLocalTimeZone, today, type DateValue } from '@internationalized/date';
	import { SvelteMap } from 'svelte/reactivity';
	import { client, type Post, type ScheduleOverview } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { getDraftPresentation } from '$lib/components/compose/draft-utils';
	import * as CalendarUi from '$lib/components/ui/calendar';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { ui } from '$lib/stores/ui.svelte';
	import { workspaceCtx } from '$lib/stores/workspace.svelte';
	import { m } from '$lib/paraglide/messages';
	import { getLocaleTag } from '$lib/i18n';
	import CalendarIcon from 'lucide-svelte/icons/calendar-days';
	import FileTextIcon from 'lucide-svelte/icons/file-text';
	import ImageIcon from 'lucide-svelte/icons/image';
	import MaximizeIcon from 'lucide-svelte/icons/maximize-2';

	let { onNavigate }: { onNavigate: (href: string) => void } = $props();

	type Publication = components['schemas']['PublicationResponse'];
	type PlannerDraft = {
		id: string;
		href: string;
		title: string;
		isThread: boolean;
		postCount: number;
		hasMedia: boolean;
		createdAt: string;
	};

	let selectedDate = $state<DateValue | undefined>(undefined);
	let calendarPlaceholder = $state<DateValue>(today(getLocalTimeZone()));
	let overview = $state<ScheduleOverview | null>(null);
	let drafts = $state.raw<PlannerDraft[]>([]);
	let loadingSchedule = $state(true);
	let loadingDrafts = $state(true);
	let overviewRequest = 0;
	let draftsRequest = 0;

	const workspaceId = $derived(workspaceCtx.currentWorkspace?.id ?? '');
	const monthString = $derived.by(() => {
		const date = calendarPlaceholder.toDate(getLocalTimeZone());
		return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
	});
	const dayCounts = $derived.by(() => {
		const counts = new SvelteMap<string, number>();
		for (const day of overview?.days ?? []) counts.set(day.date, day.count);
		return counts;
	});
	const scheduledCount = $derived(
		(overview?.days ?? []).reduce((total, day) => total + day.count, 0)
	);

	$effect(() => {
		const currentWorkspaceId = workspaceId;
		const currentMonth = monthString;
		const refresh = ui.refreshCounter;
		void refresh;
		void loadOverview(currentWorkspaceId, currentMonth);
	});

	$effect(() => {
		const currentWorkspaceId = workspaceId;
		const refresh = ui.refreshCounter;
		void refresh;
		void loadDrafts(currentWorkspaceId);
	});

	async function loadOverview(currentWorkspaceId: string, month: string) {
		const request = ++overviewRequest;
		loadingSchedule = true;
		try {
			const { data, error } = await client.GET('/posts/schedule-overview', {
				params: {
					query: {
						month,
						...(currentWorkspaceId ? { workspace_id: currentWorkspaceId } : {})
					}
				}
			});
			if (request !== overviewRequest) return;
			overview = error || !data ? null : data;
		} catch {
			if (request === overviewRequest) overview = null;
		} finally {
			if (request === overviewRequest) loadingSchedule = false;
		}
	}

	async function loadDrafts(currentWorkspaceId: string) {
		const request = ++draftsRequest;
		loadingDrafts = true;
		if (!currentWorkspaceId) {
			drafts = [];
			loadingDrafts = false;
			return;
		}

		try {
			const [legacyResult, publicationResult] = await Promise.all([
				client.GET('/posts', {
					params: {
						query: { workspace_id: currentWorkspaceId, status: 'draft', limit: 8 }
					}
				}),
				client.GET('/publications', {
					params: {
						query: {
							workspace_id: currentWorkspaceId,
							status: 'draft',
							limit: 8,
							offset: 0
						}
					}
				})
			]);
			if (request !== draftsRequest) return;
			const publications = publicationResult.error ? [] : (publicationResult.data ?? []);
			const publicationIDs = new Set(publications.map((publication) => publication.id));
			const legacyPosts = legacyResult.error ? [] : (legacyResult.data ?? []);
			drafts = [
				...publications.map(publicationDraft),
				...legacyPosts
					.filter((post) => !post.publication_id || !publicationIDs.has(post.publication_id))
					.map(legacyDraft)
			]
				.sort((left, right) => right.createdAt.localeCompare(left.createdAt))
				.slice(0, 8);
		} catch {
			if (request === draftsRequest) drafts = [];
		} finally {
			if (request === draftsRequest) loadingDrafts = false;
		}
	}

	function publicationDraft(publication: Publication): PlannerDraft {
		const segments = publication.segments ?? [];
		return {
			id: publication.id,
			href: `/publications/${encodeURIComponent(publication.id)}`,
			title:
				publication.source_text.trim() ||
				publication.title.trim() ||
				m.calendar_untitled_publication(),
			isThread: publication.intent === 'thread',
			postCount: Math.max(1, segments.length),
			hasMedia:
				(publication.media?.length ?? 0) > 0 ||
				segments.some((segment) => (segment.media?.length ?? 0) > 0),
			createdAt: publication.created_at
		};
	}

	function legacyDraft(post: Post): PlannerDraft {
		const presentation = getDraftPresentation(post);
		return {
			id: post.id,
			href: `/posts/${post.id}`,
			title: presentation.title,
			isThread: presentation.isThread,
			postCount: presentation.postCount,
			hasMedia: presentation.hasMedia,
			createdAt: post.created_at
		};
	}

	function handleDateChange(date: DateValue | undefined) {
		selectedDate = undefined;
		if (date) ui.openDayPosts(date);
	}

	type DayMarkerArgs = { day: DateValue; outsideMonth: boolean };
</script>

{#snippet dayMarker({ day, outsideMonth }: DayMarkerArgs)}
	{@const count = dayCounts.get(day.toString()) ?? 0}
	<div class="relative flex size-(--cell-size) items-center justify-center">
		<CalendarUi.Day />
		{#if !outsideMonth && count > 0}
			<span
				class="pointer-events-none absolute bottom-0.5 size-1 rounded-full bg-primary ring-1 ring-sidebar"
				aria-hidden="true"
			></span>
		{/if}
	</div>
{/snippet}

<div class="flex shrink-0 flex-col" data-testid="desktop-sidebar-planner">
	<section class="border-b border-sidebar-border px-2 pb-3">
		<div class="mb-1 flex h-8 items-center justify-between px-2">
			<button
				type="button"
				class="group inline-flex items-center gap-1.5 text-xs font-medium text-sidebar-foreground focus-visible:ring-2 focus-visible:ring-sidebar-ring focus-visible:outline-none"
				onclick={() => onNavigate('/calendar')}
				aria-label={m.sidebar_calendar()}
			>
				<CalendarIcon class="size-3.5 text-primary" />
				<span>{m.sidebar_calendar()}</span>
				<MaximizeIcon
					class="size-3 text-sidebar-foreground/42 transition-colors group-hover:text-sidebar-foreground"
				/>
			</button>
			<span class="text-xs text-sidebar-foreground/48 tabular-nums">
				{#if loadingSchedule}{m.sidebar_schedule_loading()}{:else}{m.sidebar_schedule_count({
						count: scheduledCount
					})}{/if}
			</span>
		</div>

		<CalendarUi.Calendar
			type="single"
			bind:value={selectedDate}
			bind:placeholder={calendarPlaceholder}
			onValueChange={handleDateChange}
			day={dayMarker}
			locale={getLocaleTag()}
			weekStartsOn={workspaceCtx.settings.week_start as 0 | 1 | 2 | 3 | 4 | 5 | 6}
			class="w-full bg-transparent p-0 select-none [--cell-size:1.75rem] [&_[role=gridcell]_[role=button][data-today]]:bg-sidebar-primary [&_[role=gridcell]_[role=button][data-today]]:text-sidebar-primary-foreground [&_tr]:justify-between"
		/>
	</section>

	<section class="min-h-0 border-b border-sidebar-border px-2 py-3">
		<div class="mb-1 flex h-7 items-center justify-between px-2">
			<div class="flex items-center gap-1.5">
				<span class="text-xs font-medium tracking-[0.1em] text-sidebar-foreground/52 uppercase"
					>{m.sidebar_drafts()}</span
				>
				{#if !loadingDrafts && drafts.length > 0}
					<span class="text-xs text-sidebar-foreground/38 tabular-nums">{drafts.length}</span>
				{/if}
			</div>
			<button
				type="button"
				class="rounded-sm px-1.5 py-1 text-xs font-medium text-sidebar-foreground/58 hover:bg-sidebar-accent hover:text-sidebar-foreground focus-visible:ring-2 focus-visible:ring-sidebar-ring focus-visible:outline-none"
				onclick={() => onNavigate('/activity?tab=drafts')}
			>
				{m.sidebar_view_all()}
			</button>
		</div>

		{#if loadingDrafts}
			<div class="space-y-1 px-1 py-1" aria-label={m.sidebar_drafts_loading()}>
				{#each [1, 2, 3] as placeholder (placeholder)}
					<div class="flex h-9 items-center gap-2 px-1.5">
						<Skeleton class="size-6 rounded-md" />
						<Skeleton class="h-3 flex-1" />
					</div>
				{/each}
			</div>
		{:else if drafts.length === 0}
			<button
				type="button"
				class="flex w-full items-start gap-2 rounded-md px-2 py-2.5 text-left hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-sidebar-ring focus-visible:outline-none"
				onclick={() => onNavigate('/')}
			>
				<FileTextIcon class="mt-0.5 size-3.5 shrink-0 text-sidebar-foreground/38" />
				<span class="text-xs leading-4 text-sidebar-foreground/52"
					>{m.sidebar_drafts_autosave_empty()}</span
				>
			</button>
		{:else}
			<ul class="max-h-40 space-y-0.5 overflow-y-auto" data-testid="sidebar-draft-list">
				{#each drafts as draft (draft.id)}
					<li>
						<button
							type="button"
							class="group flex min-h-9 w-full items-center gap-2 rounded-md px-2 py-1.5 text-left hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-sidebar-ring focus-visible:outline-none"
							onclick={() => onNavigate(draft.href)}
							aria-label={m.sidebar_resume_draft({ title: draft.title })}
						>
							<span
								class="flex size-6 shrink-0 items-center justify-center rounded-md bg-sidebar-accent/70 text-sidebar-foreground/58 group-hover:text-sidebar-foreground"
							>
								<FileTextIcon class="size-3.5" />
							</span>
							<span class="min-w-0 flex-1">
								<span class="block truncate text-xs font-medium text-sidebar-foreground/88"
									>{draft.title}</span
								>
								{#if draft.isThread}
									<span class="block text-xs leading-4 text-sidebar-foreground/45"
										>{m.sidebar_thread_count({ count: draft.postCount })}</span
									>
								{/if}
							</span>
							{#if draft.hasMedia}
								<ImageIcon
									class="size-3 shrink-0 text-sidebar-foreground/38"
									aria-label={m.sidebar_has_media()}
								/>
							{/if}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</div>
