import { useSidebar } from '@modules/ui/sidebar';
import { cn } from '@modules/utils';
import { Info } from 'lucide-react';

export const AddProjectDialogInfo = () => {
	const { state, isMobile } = useSidebar();
	const show = state !== 'collapsed' || isMobile;
	return (
		<div className={cn('w-full bg-muted rounded-sm p-2 flex items-start gap-2 mb-2')} hidden={!show}>
			<Info className="w-4 h-4 mt-[2px] flex-shrink-0" />
			<p className="text-[10px] leading-snug">
				<span className="font-bold">Got a token launch coming up? </span>
				Share your moon date and let early users discover your project before takeoff.
			</p>
		</div>
	);
};
