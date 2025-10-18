import z from 'zod';

export const MoonMapMenu = {
	PROJECTS: '/projects',
	SPHERES: '/spheres',
	NEBULAS: '/nebulas',
	WAVES: '/waves',
} as const;
export const MoonMapMenuSchema = z.enum([...Object.keys(MoonMapMenu)] as [MoonMapMenuType, ...MoonMapMenuType[]]);

export type MoonMapMenuType = keyof typeof MoonMapMenu;
export type MoonMapMenuUrl = (typeof MoonMapMenu)[MoonMapMenuType];
