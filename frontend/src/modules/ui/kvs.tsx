// modules/ui/kvs.tsx
export function KVS({ items }: { items: Array<{ k: string; v: React.ReactNode }> }) {
	return (
		<dl className="grid grid-cols-3 gap-y-2 text-sm">
			{items.map(({ k, v }) => (
				<div key={k} className="contents">
					<dt className="font-medium text-muted-foreground col-span-1">{k}</dt>
					<dd className="truncate col-span-2">{v}</dd>
				</div>
			))}
		</dl>
	);
}
