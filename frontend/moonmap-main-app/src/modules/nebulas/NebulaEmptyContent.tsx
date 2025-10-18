import { Telescope, Construction } from 'lucide-react';

export const NebulaEmptyContent = () => {
	return (
		<div className="flex flex-col h-full overflow-y-auto p-4 py-0">
			<div className="flex flex-col items-center justify-center text-center rounded-2xl border border-dashed border-muted-foreground/20 bg-muted/40 p-8 h-full">
				<Telescope strokeWidth={0.9} className="size-30 text-muted-foreground mb-3" />
				<h2 className="text-lg font-semibold text-foreground">Nebulas are coming soon</h2>
				<p className="mt-2 text-sm text-muted-foreground max-w-md">
					Nebulas will be live spaces for video conferencing and collaboration. We’re still building this feature — stay tuned.
				</p>

				<p className="text-sm text-muted-foreground max-w-md mt-4 flex-row flex items-center justify-center">
					<Construction className="size-4 mr-2" />
					<span className="text-left">Work in progress — Nebulas will be available soon ✨</span>
				</p>
			</div>
		</div>
	);
};
