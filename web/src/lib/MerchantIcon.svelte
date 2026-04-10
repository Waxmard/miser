<script lang="ts">
	import type { SimpleIcon } from 'simple-icons';
	import {
		siJetblue,
		siVerizon,
		siNetflix,
		siSpotify,
		siApple,
		siUber,
		siLyft,
		siDoordash,
		siVenmo,
		siPaypal,
		siDelta,
		siAmericanairlines,
		siSouthwestairlines,
		siTarget,
		siStarbucks,
		siGoogle,
		siMcdonalds,
		siUnitedairlines,
		siMetro,
		siYoutube,
		siTwitch,
		siZoom,
		siDropbox,
		siGithub,
		siEtsy,
		siEbay,
		siHilton,
		siMarriott,
		siAirbnb,
		siExpedia,
		siBankofamerica,
		siWellsfargo,
		siUsps,
		siFedex,
		siUps,
		siDhl,
		siTesla,
		siFord,
		siToyota,
		siHyundai,
		siChevrolet,
		siChase,
		siBookingdotcom
	} from 'simple-icons';

	export let merchant: string;
	export let size: number = 32;

	// Each entry: [keyword-to-match-against-normalized-name, icon]
	// Ordered so more-specific strings come first
	const iconMap: [string, SimpleIcon][] = [
		['american airlines', siAmericanairlines],
		['southwest airlines', siSouthwestairlines],
		['southwest', siSouthwestairlines],
		['united airlines', siUnitedairlines],
		['bank of america', siBankofamerica],
		['wells fargo', siWellsfargo],
		['booking.com', siBookingdotcom],
		['jetblue', siJetblue],
		['verizon', siVerizon],
		['netflix', siNetflix],
		['spotify', siSpotify],
		['apple', siApple],
		['uber', siUber],
		['lyft', siLyft],
		['doordash', siDoordash],
		['venmo', siVenmo],
		['paypal', siPaypal],
		['delta', siDelta],
		['target', siTarget],
		['starbucks', siStarbucks],
		['google', siGoogle],
		['mcdonalds', siMcdonalds],
		["mcdonald's", siMcdonalds],
		['metro', siMetro],
		['youtube', siYoutube],
		['twitch', siTwitch],
		['zoom', siZoom],
		['dropbox', siDropbox],
		['github', siGithub],
		['etsy', siEtsy],
		['ebay', siEbay],
		['hilton', siHilton],
		['marriott', siMarriott],
		['airbnb', siAirbnb],
		['expedia', siExpedia],
		['usps', siUsps],
		['fedex', siFedex],
		['ups', siUps],
		['dhl', siDhl],
		['tesla', siTesla],
		['ford', siFord],
		['toyota', siToyota],
		['hyundai', siHyundai],
		['chevrolet', siChevrolet],
		['chevy', siChevrolet],
		['chase', siChase]
	];

	// Return false if hex color is too light to see on our cream background
	function isUsable(hex: string): boolean {
		const r = parseInt(hex.slice(0, 2), 16) / 255;
		const g = parseInt(hex.slice(2, 4), 16) / 255;
		const b = parseInt(hex.slice(4, 6), 16) / 255;
		// Relative luminance (WCAG formula)
		const lum = 0.2126 * r + 0.7152 * g + 0.0722 * b;
		return lum < 0.88;
	}

	const normalized = merchant.toLowerCase();

	const matchedIcon = iconMap.find(([kw]) => normalized.includes(kw))?.[1] ?? null;
	const iconColor =
		matchedIcon && isUsable(matchedIcon.hex) ? `#${matchedIcon.hex}` : 'var(--color-text-muted)';

	// Fallback: deterministic letter avatar
	const firstLetter = merchant.trim().charAt(0).toUpperCase();
	const AVATAR_COLORS = [
		'#6B7280', // slate
		'#9A7B5A', // warm tan
		'#5B7A6B', // muted teal
		'#7C6B8A', // dusty purple
		'#8A6B6B', // mauve
		'#6B7C8A', // steel
		'#8A7C6B', // warm gray
		'#7A8A6B' // sage
	];
	function avatarColor(s: string): string {
		let h = 0;
		for (let i = 0; i < s.length; i++) {
			h = (s.charCodeAt(i) + ((h << 5) - h)) | 0;
		}
		return AVATAR_COLORS[Math.abs(h) % AVATAR_COLORS.length];
	}
	const bg = avatarColor(merchant);

	// Avatar font size: ~44% of container size
	$: avatarFontSize = Math.round(size * 0.44);
</script>

{#if matchedIcon}
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
			<path d={matchedIcon.path} />
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
