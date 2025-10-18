import { useAppStore } from '@modules/stores/main/app-hooks';
import { Tooltip, TooltipContent, TooltipTrigger } from '@modules/ui/tooltip';
import { X } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useRemoveProjectLogo } from './useUploadLogo';
import { useFormContext } from 'react-hook-form';
import { CreateProjectInput } from './DialogBody';

export function RemoveCoinButton() {
	const form = useFormContext<CreateProjectInput>();
	const [open, setOpen] = useState(false);
	const projectId = useAppStore((state) => state.draftProject?.id);
	const removeProjectLogo = useRemoveProjectLogo();
	const removeDraftProjectLogo = useAppStore((state) => state.removeDraftProjectLogo);

	const onRemoveLogo = useCallback(() => {
		if (!projectId) return;
		form.setValue('imageUrl', '');
		removeDraftProjectLogo();
		removeProjectLogo(projectId);
	}, [form, projectId, removeDraftProjectLogo, removeProjectLogo]);

	return (
		<div className="absolute right-1 top-1 z-50">
			<Tooltip open={open}>
				<TooltipTrigger
					asChild
					onMouseEnter={() => setOpen(true)}
					onMouseLeave={() => setOpen(false)}
					onFocus={() => setOpen(false)}
					onBlur={() => setOpen(false)}
				>
					<button className="w-6 h-6 flex items-center justify-center bg-primary rounded-full" onClick={onRemoveLogo}>
						<X className="text-primary-foreground size-4" />
					</button>
				</TooltipTrigger>
				<TooltipContent>Remove coin image</TooltipContent>
			</Tooltip>
		</div>
	);
}
