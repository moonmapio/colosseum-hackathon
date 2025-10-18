export interface PageProps {
	searchParams: Promise<{ token?: string }>;
	params: Promise<{ language: string }>;
}
