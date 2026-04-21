<script lang="ts">
	import { parseIconSlug } from '$lib/icons';
	import { getSimpleIcon, type SlimSimpleIcon } from '$lib/simpleIconCatalog';

	export let name: string;
	export let size: number = 32;
	export let iconSlug: string | null = null;

		type Resolved =
			| { type: 'si'; icon: SlimSimpleIcon }
			| { type: 'emoji'; emoji: string }
			| { type: 'none' };

	function isUsable(hex: string): boolean {
		const r = parseInt(hex.slice(0, 2), 16) / 255;
		const g = parseInt(hex.slice(2, 4), 16) / 255;
		const b = parseInt(hex.slice(4, 6), 16) / 255;
		return 0.2126 * r + 0.7152 * g + 0.0722 * b < 0.88;
	}

	let resolved: Resolved = { type: 'none' };
	$: void resolve(iconSlug);
	async function resolve(slug: string | null) {
		if (!slug) {
			resolved = { type: 'none' };
			return;
		}
		const { library, name: parsed } = parseIconSlug(slug);
		if (library === 'emoji') {
			resolved = { type: 'emoji', emoji: parsed };
			return;
		}
		const icon = await getSimpleIcon(parsed);
		if (parseIconSlug(iconSlug ?? '').name !== parsed) return;
		resolved = icon ? { type: 'si', icon } : { type: 'none' };
	}
	$: siIconColor =
		resolved.type === 'si' && isUsable(resolved.icon.hex)
			? `#${resolved.icon.hex}`
			: 'var(--color-text-muted)';

	const AVATAR_COLORS = [
		'#6B7280', '#9A7B5A', '#5B7A6B', '#7C6B8A',
		'#8A6B6B', '#6B7C8A', '#8A7C6B', '#7A8A6B'
	];
	function avatarColor(s: string): string {
		let h = 0;
		for (let i = 0; i < s.length; i++) h = (s.charCodeAt(i) + ((h << 5) - h)) | 0;
		return AVATAR_COLORS[Math.abs(h) % AVATAR_COLORS.length];
	}

	$: firstLetter = name.trim().charAt(0).toUpperCase();
	$: bg = avatarColor(name);
	$: avatarFontSize = Math.round(size * 0.44);
	$: emojiFontSize = Math.round(size * 0.6);
</script>

<button
	class="icon-btn"
	class:avatar={resolved.type === 'none'}
	style="width:{size}px;height:{size}px;{resolved.type === 'none' ? `background:${bg};font-size:${avatarFontSize}px` : ''}"
	title="Change {name} icon"
	on:click
>
	{#if resolved.type === 'si'}
		<span style="color:{siIconColor};display:flex;align-items:center;justify-content:center">
			<svg
				role="img"
				viewBox="0 0 24 24"
				width={size * 0.6}
				height={size * 0.6}
				fill="currentColor"
				aria-label={name}
			>
				<path d={resolved.icon.path} />
			</svg>
		</span>
	{:else if resolved.type === 'emoji'}
		<span style="font-size:{emojiFontSize}px;line-height:1" aria-hidden="true">
			{resolved.emoji}
		</span>
	{:else}
		{firstLetter}
	{/if}
</button>

<style>
	.icon-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius);
		background: var(--color-surface-alt);
		border: 1px solid var(--color-border);
		flex-shrink: 0;
		cursor: pointer;
		padding: 0;
		transition: opacity 0.12s, border-color 0.12s;
	}

	.icon-btn:hover {
		opacity: 0.75;
		border-color: var(--color-accent);
	}

	.icon-btn.avatar {
		border-color: transparent;
		color: #ffffff;
		font-weight: 600;
		font-family: var(--font-sans);
		letter-spacing: 0;
	}
</style>
