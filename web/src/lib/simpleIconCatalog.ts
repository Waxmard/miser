import type { SimpleIcon } from 'simple-icons';

let catalogPromise: Promise<Map<string, SimpleIcon>> | null = null;

function loadCatalog(): Promise<Map<string, SimpleIcon>> {
	if (!catalogPromise) {
		catalogPromise = import('simple-icons').then((mod) => {
			const map = new Map<string, SimpleIcon>();
			for (const icon of Object.values(mod) as SimpleIcon[]) {
				if (icon && typeof icon === 'object' && 'slug' in icon) {
					map.set(icon.slug, icon);
				}
			}
			return map;
		});
	}
	return catalogPromise;
}

export async function getSimpleIcon(slug: string): Promise<SimpleIcon | null> {
	const catalog = await loadCatalog();
	return catalog.get(slug) ?? null;
}
