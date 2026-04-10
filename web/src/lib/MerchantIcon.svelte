<script lang="ts">
	import * as si from 'simple-icons';
	import type { SimpleIcon } from 'simple-icons';

	export let merchant: string;
	export let size: number = 32;
	/** Explicit icon slug override from the DB — takes priority over the built-in keyword map. */
	export let iconSlug: string | null = null;

	// Build slug → icon lookup once at module load
	const bySlug = new Map<string, SimpleIcon>(
		Object.values(si).map((icon) => [icon.slug, icon])
	);

	// Built-in keyword → slug mapping for common brands (fallback when no DB override)
	const keywordMap: [string, string][] = [
		['american airlines', 'americanairlines'],
		['southwest airlines', 'southwestairlines'],
		['southwest', 'southwestairlines'],
		['united airlines', 'unitedairlines'],
		['bank of america', 'bankofamerica'],
		['wells fargo', 'wellsfargo'],
		['booking.com', 'bookingdotcom'],
		['jetblue', 'jetblue'],
		['verizon', 'verizon'],
		['netflix', 'netflix'],
		['spotify', 'spotify'],
		['apple', 'apple'],
		['uber', 'uber'],
		['lyft', 'lyft'],
		['doordash', 'doordash'],
		['venmo', 'venmo'],
		['paypal', 'paypal'],
		['delta', 'delta'],
		['target', 'target'],
		['starbucks', 'starbucks'],
		['google', 'google'],
		['mcdonalds', 'mcdonalds'],
		["mcdonald's", 'mcdonalds'],
		['metro', 'metro'],
		['youtube', 'youtube'],
		['twitch', 'twitch'],
		['zoom', 'zoom'],
		['dropbox', 'dropbox'],
		['github', 'github'],
		['etsy', 'etsy'],
		['ebay', 'ebay'],
		['hilton', 'hilton'],
		['marriott', 'marriott'],
		['airbnb', 'airbnb'],
		['expedia', 'expedia'],
		['usps', 'usps'],
		['fedex', 'fedex'],
		['ups', 'ups'],
		['dhl', 'dhl'],
		['tesla', 'tesla'],
		['ford', 'ford'],
		['toyota', 'toyota'],
		['hyundai', 'hyundai'],
		['chevrolet', 'chevrolet'],
		['chevy', 'chevrolet'],
		['chase', 'chase']
	];

	function isUsable(hex: string): boolean {
		const r = parseInt(hex.slice(0, 2), 16) / 255;
		const g = parseInt(hex.slice(2, 4), 16) / 255;
		const b = parseInt(hex.slice(4, 6), 16) / 255;
		return 0.2126 * r + 0.7152 * g + 0.0722 * b < 0.88;
	}

	function resolveIcon(slug: string | null, merchantName: string): SimpleIcon | null {
		if (slug) return bySlug.get(slug) ?? null;
		const norm = merchantName.toLowerCase();
		const matched = keywordMap.find(([kw]) => norm.includes(kw));
		if (matched) return bySlug.get(matched[1]) ?? null;
		return null;
	}

	$: icon = resolveIcon(iconSlug, merchant);
	$: iconColor = icon && isUsable(icon.hex) ? `#${icon.hex}` : 'var(--color-text-muted)';

	// Fallback letter avatar
	const firstLetter = merchant.trim().charAt(0).toUpperCase();
	const AVATAR_COLORS = [
		'#6B7280',
		'#9A7B5A',
		'#5B7A6B',
		'#7C6B8A',
		'#8A6B6B',
		'#6B7C8A',
		'#8A7C6B',
		'#7A8A6B'
	];
	function avatarColor(s: string): string {
		let h = 0;
		for (let i = 0; i < s.length; i++) h = (s.charCodeAt(i) + ((h << 5) - h)) | 0;
		return AVATAR_COLORS[Math.abs(h) % AVATAR_COLORS.length];
	}
	const bg = avatarColor(merchant);
	$: avatarFontSize = Math.round(size * 0.44);
</script>

{#if icon}
	<span
		class="icon-wrap"
		style="width:{size}px;height:{size}px;color:{iconColor}"
		title={merchant}
	>
		<svg
			role="img"
			viewBox="0 0 24 24"
			width={size * 0.6}
			height={size * 0.6}
			fill="currentColor"
			aria-label={merchant}
		>
			<path d={icon.path} />
		</svg>
	</span>
{:else}
	<span
		class="avatar"
		style="width:{size}px;height:{size}px;background:{bg};font-size:{avatarFontSize}px"
		aria-hidden="true"
		title={merchant}
	>
		{firstLetter}
	</span>
{/if}

<style>
	.icon-wrap {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius);
		background: var(--color-surface-alt);
		flex-shrink: 0;
		border: 1px solid var(--color-border);
	}

	.avatar {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius);
		color: #ffffff;
		font-weight: 600;
		flex-shrink: 0;
		font-family: var(--font-sans);
		letter-spacing: 0;
	}
</style>
