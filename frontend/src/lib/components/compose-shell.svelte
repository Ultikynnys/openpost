<script lang="ts">
	import ComposeSimple from './compose-simple.svelte';
	import ComposeFocusedPublication from './compose-focused-publication.svelte';
	import { Button } from '$lib/components/ui/button';
	import { COMPOSER_MODES, isLegacyComposerMode, type ComposerModeKey } from './compose/modes';
	import AlignLeftIcon from 'lucide-svelte/icons/align-left';
	import ImageIcon from 'lucide-svelte/icons/image';
	import ImagesIcon from 'lucide-svelte/icons/images';
	import LinkIcon from 'lucide-svelte/icons/link';
	import ListIcon from 'lucide-svelte/icons/list';
	import PlayIcon from 'lucide-svelte/icons/play';
	import SmartphoneIcon from 'lucide-svelte/icons/smartphone';
	import VideoIcon from 'lucide-svelte/icons/video';

	const modeIcons: Partial<Record<ComposerModeKey, typeof AlignLeftIcon>> = {
		short_text: AlignLeftIcon,
		thread: ListIcon,
		link_share: LinkIcon,
		image_post: ImageIcon,
		carousel: ImagesIcon,
		story: SmartphoneIcon,
		short_video: PlayIcon,
		long_video: VideoIcon
	};

	let selectedMode = $state<ComposerModeKey>('short_text');
	let lastComposerThreadState = false;

	function handleThreadStateChange(isThread: boolean) {
		if (isThread) {
			selectedMode = 'thread';
		} else if (lastComposerThreadState && selectedMode === 'thread') {
			selectedMode = 'short_text';
		}
		lastComposerThreadState = isThread;
	}
</script>

<div class="flex min-h-0 flex-1 flex-col bg-background" data-testid="compose-shell">
	<div class="border-b bg-background/95 px-3 py-2 md:px-4">
		<div class="flex min-w-0 items-center gap-2 overflow-x-auto" data-testid="composer-mode-picker">
			{#each COMPOSER_MODES as mode (mode.key)}
				{@const Icon = modeIcons[mode.key] ?? AlignLeftIcon}
				<Button
					type="button"
					variant={selectedMode === mode.key ? 'default' : 'ghost'}
					size="sm"
					class="h-8 shrink-0 gap-1.5 px-2.5 text-xs"
					aria-pressed={selectedMode === mode.key}
					onclick={() => (selectedMode = mode.key)}
				>
					<Icon class="h-3.5 w-3.5" />
					{mode.label}
				</Button>
			{/each}
		</div>
	</div>

	{#if isLegacyComposerMode(selectedMode)}
		<div data-testid="legacy-composer-shell" class="flex min-h-0 flex-1 flex-col">
			<ComposeSimple onThreadStateChange={handleThreadStateChange} />
		</div>
	{:else}
		{#key selectedMode}
			<ComposeFocusedPublication mode={selectedMode} />
		{/key}
	{/if}
</div>
