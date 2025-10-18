'use client';

import { useAppStore } from '@modules/stores/main/app-hooks';
import { Video } from 'lucide-react';
import { NebulaEmptyContent } from './NebulaEmptyContent';
import { MoonMapContentLayout } from '@modules/layout/MoonMapContentLayout';

export const Nebulas = () => {
	const selectedMenu = useAppStore((state) => state.selectedMenu);
	const show = selectedMenu === 'NEBULAS';
	if (!show) return null;

	return (
		<MoonMapContentLayout>
			{/* header pequeño arriba */}
			<div className="flex items-center px-4 shrink-0">
				<Video className="w-6 h-6 mr-2" />
				<h1 className="text-md font-medium ml-1 text-muted-foreground">Nebulas</h1>
			</div>

			<NebulaEmptyContent />
			{/* <SphereEmptyContent /> */}
		</MoonMapContentLayout>
	);
};
