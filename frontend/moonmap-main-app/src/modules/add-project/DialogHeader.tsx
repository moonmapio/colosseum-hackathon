import { DialogDescription, DialogTitle, DialogHeader as UiDialogHeader } from '@modules/ui/dialog';
import { Sparkles } from 'lucide-react';

export function DialogHeader() {
	return (
		<UiDialogHeader className="mb-3">
			<DialogTitle className="flex flex-row items-center">
				<span>
					<Sparkles size={26} className="mr-2" />
				</span>
				New Project
			</DialogTitle>
			<DialogDescription className="text-xs">
				Add your coin to <span className="font-bold underline underline-offset-2 decoration-primary text-primary">MoonMap</span>, build your
				community, spark the hype, and steer your moonshot. Drop the alpha, rally the degens, and we’ll amplify your wins across the feed!
				So no one on CT misses it. WAGMI.
			</DialogDescription>
		</UiDialogHeader>
	);
}
