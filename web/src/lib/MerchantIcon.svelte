<script lang="ts">
	import type { SimpleIcon } from 'simple-icons';
	import { parseIconSlug } from '$lib/icons';
	import { getSimpleIcon } from '$lib/simpleIconCatalog';

	export let merchant: string;
	export let size: number = 32;
	/** Explicit icon slug override from the DB — takes priority over the built-in keyword map. */
	export let iconSlug: string | null = null;

	// Built-in keyword → slug mapping for common brands (fallback when no DB override)
	const keywordMap: [string, string][] = [
		// Airlines
		['american airlines', 'americanairlines'],
		['southwest airlines', 'southwestairlines'],
		['southwest', 'southwestairlines'],
		['united airlines', 'unitedairlines'],
		['alaska airlines', 'alaskaairlines'],
		['jetblue', 'jetblue'],
		['delta', 'delta'],
		['spirit airlines', 'spiritairlines'],
		['frontier airlines', 'frontierairlines'],
		['british airways', 'britishairways'],
		['air france', 'airfrance'],
		['lufthansa', 'lufthansa'],
		['emirates', 'emirates'],

		// Hotels
		['hilton', 'hilton'],
		['marriott', 'marriott'],
		['hyatt', 'hyatt'],
		['airbnb', 'airbnb'],
		['expedia', 'expedia'],
		['booking.com', 'bookingdotcom'],
		['vrbo', 'vrbo'],

		// Banks & Finance
		['bank of america', 'bankofamerica'],
		['wells fargo', 'wellsfargo'],
		['chase', 'chase'],
		['american express', 'americanexpress'],
		['amex', 'americanexpress'],
		['capital one', 'capitalone'],
		['citibank', 'citibank'],
		['citi ', 'citibank'],
		['discover', 'discover'],
		['ally', 'ally'],
		['schwab', 'charlesschwab'],
		['charles schwab', 'charlesschwab'],
		['fidelity', 'fidelity'],
		['vanguard', 'vanguard'],
		['robinhood', 'robinhood'],
		['coinbase', 'coinbase'],
		['paypal', 'paypal'],
		['venmo', 'venmo'],
		['cashapp', 'cashapp'],
		['cash app', 'cashapp'],
		['zelle', 'zelle'],
		['sofi', 'sofi'],

		// Streaming & Subscriptions
		['netflix', 'netflix'],
		['spotify', 'spotify'],
		['hulu', 'hulu'],
		['disney', 'disneyplus'],
		['hbo', 'hbo'],
		['paramount', 'paramount'],
		['peacock', 'peacocktv'],
		['apple tv', 'appletv'],
		['amazon prime', 'prime'],
		['youtube', 'youtube'],
		['twitch', 'twitch'],
		['pandora', 'pandora'],
		['tidal', 'tidal'],
		['audible', 'audible'],
		['kindle', 'kindle'],

		// Tech & Software
		['apple', 'apple'],
		['google', 'google'],
		['microsoft', 'microsoft'],
		['adobe', 'adobe'],
		['amazon', 'amazon'],
		['slack', 'slack'],
		['zoom', 'zoom'],
		['dropbox', 'dropbox'],
		['github', 'github'],
		['openai', 'openai'],
		['claude', 'anthropic'],
		['figma', 'figma'],
		['notion', 'notion'],
		['shopify', 'shopify'],
		['squarespace', 'squarespace'],

		// Social / Gig
		['uber', 'uber'],
		['lyft', 'lyft'],
		['doordash', 'doordash'],
		['grubhub', 'grubhub'],
		['instacart', 'instacart'],
		['postmates', 'postmates'],
		['etsy', 'etsy'],
		['ebay', 'ebay'],
		['facebook', 'facebook'],
		['instagram', 'instagram'],
		['tiktok', 'tiktok'],
		['twitter', 'twitter'],
		['linkedin', 'linkedin'],
		['reddit', 'reddit'],
		['pinterest', 'pinterest'],
		['snapchat', 'snapchat'],
		['discord', 'discord'],

		// Retail & Grocery
		['target', 'target'],
		['walmart', 'walmart'],
		['costco', 'costco'],
		['whole foods', 'wholefoods'],
		['whole fds', 'wholefoods'],
		['trader joe', 'traderjoes'],
		['kroger', 'kroger'],
		['safeway', 'safeway'],
		['publix', 'publix'],
		['ikea', 'ikea'],
		['wayfair', 'wayfair'],
		['home depot', 'homedepot'],
		['homedepot', 'homedepot'],
		["lowe's", 'lowes'],
		['lowes', 'lowes'],
		['best buy', 'bestbuy'],
		['bestbuy', 'bestbuy'],
		['chewy', 'chewy'],
		['petco', 'petco'],
		['petsmart', 'petsmart'],
		['rei ', 'rei'],
		['sephora', 'sephora'],
		['ulta', 'ulta'],
		['nike', 'nike'],
		['adidas', 'adidas'],

		// Food & Dining
		['starbucks', 'starbucks'],
		['mcdonalds', 'mcdonalds'],
		["mcdonald's", 'mcdonalds'],
		['chipotle', 'chipotle'],
		['panera', 'panera'],
		['dunkin', 'dunkindonuts'],
		['subway', 'subway'],
		['dominos', 'dominospizza'],
		["domino's", 'dominospizza'],
		['chick-fil-a', 'chickfila'],
		['taco bell', 'tacobell'],
		["wendy's", 'wendys'],
		['wendys', 'wendys'],
		["burger king", 'burgerking'],
		['five guys', 'fiveguys'],
		['shake shack', 'shakeshack'],
		['in-n-out', 'innout'],

		// Telecom
		['verizon', 'verizon'],
		['t-mobile', 'tmobile'],
		['tmobile', 'tmobile'],
		['at&t', 'att'],
		['att', 'att'],
		['xfinity', 'xfinity'],
		['comcast', 'xfinity'],

		// Gas & Auto
		['shell', 'shell'],
		['chevron', 'chevron'],
		['bp', 'bp'],
		['exxon', 'exxonmobil'],
		['mobil', 'exxonmobil'],
		['tesla', 'tesla'],
		['ford', 'ford'],
		['toyota', 'toyota'],
		['hyundai', 'hyundai'],
		['chevrolet', 'chevrolet'],
		['chevy', 'chevrolet'],
		['volkswagen', 'volkswagen'],
		['bmw', 'bmw'],
		['honda', 'honda'],
		['nissan', 'nissan'],
		['geico', 'geico'],
		['progressive', 'progressive'],
		['state farm', 'statefarm'],

		// Shipping
		['usps', 'usps'],
		['fedex', 'fedex'],
		['ups', 'ups'],
		['dhl', 'dhl'],

		// Healthcare / Pharmacy
		['cvs', 'cvs'],
		['walgreens', 'walgreens'],

		// Misc
		['metro', 'metro'],
		['amtrak', 'amtrak'],
	];

	type Resolved =
		| { type: 'si'; icon: SimpleIcon }
		| { type: 'emoji'; emoji: string }
		| { type: 'none' };

	function isUsable(hex: string): boolean {
		const r = parseInt(hex.slice(0, 2), 16) / 255;
		const g = parseInt(hex.slice(2, 4), 16) / 255;
		const b = parseInt(hex.slice(4, 6), 16) / 255;
		return 0.2126 * r + 0.7152 * g + 0.0722 * b < 0.88;
	}

	let resolved: Resolved = { type: 'none' };
	let resolveToken = 0;
	$: void resolve(iconSlug, merchant);
	async function resolve(slug: string | null, merchantName: string) {
		const token = ++resolveToken;
		if (slug) {
			const { library, name } = parseIconSlug(slug);
			if (library === 'emoji') {
				resolved = { type: 'emoji', emoji: name };
				return;
			}
			const icon = await getSimpleIcon(name);
			if (token !== resolveToken) return;
			resolved = icon ? { type: 'si', icon } : { type: 'none' };
			return;
		}
		const norm = merchantName.toLowerCase();
		const matched = keywordMap.find(([kw]) => norm.includes(kw));
		if (!matched) {
			resolved = { type: 'none' };
			return;
		}
		const icon = await getSimpleIcon(matched[1]);
		if (token !== resolveToken) return;
		resolved = icon ? { type: 'si', icon } : { type: 'none' };
	}
	$: siIconColor =
		resolved.type === 'si' && isUsable(resolved.icon.hex)
			? `#${resolved.icon.hex}`
			: 'var(--color-text-muted)';

	// Fallback letter avatar
	$: firstLetter = merchant.trim().charAt(0).toUpperCase();
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
	$: bg = avatarColor(merchant);
	$: avatarFontSize = Math.round(size * 0.44);
	$: emojiFontSize = Math.round(size * 0.6);
</script>

{#if resolved.type === 'si'}
	<span
		class="icon-wrap"
		style="width:{size}px;height:{size}px;color:{siIconColor}"
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
			<path d={resolved.icon.path} />
		</svg>
	</span>
{:else if resolved.type === 'emoji'}
	<span
		class="icon-wrap emoji-wrap"
		style="width:{size}px;height:{size}px;font-size:{emojiFontSize}px"
		role="img"
		aria-label={merchant}
		title={merchant}
	>
		{resolved.emoji}
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

	.emoji-wrap {
		line-height: 1;
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
