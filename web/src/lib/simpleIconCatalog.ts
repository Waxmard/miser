export type SlimSimpleIcon = {
	slug: string;
	hex: string;
	path: string;
};

type SlimSimpleIconRecord = {
	hex: string;
	path: string;
};

type SlimSimpleIconModule = {
	default: SlimSimpleIconRecord;
};

const iconLoaders = import.meta.glob('./generated/simple-icons/*.json');
const iconPromiseCache = new Map<string, Promise<SlimSimpleIcon | null>>();

export async function getSimpleIcon(slug: string): Promise<SlimSimpleIcon | null> {
	const cached = iconPromiseCache.get(slug);
	if (cached) {
		return cached;
	}

	const key = `./generated/simple-icons/${slug}.json`;
	const load = iconLoaders[key] as (() => Promise<SlimSimpleIconModule>) | undefined;
	if (!load) {
		return null;
	}

	const promise = load().then((mod) => ({ slug, ...mod.default }));
	iconPromiseCache.set(slug, promise);
	return promise;
}
