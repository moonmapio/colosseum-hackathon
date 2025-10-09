'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { ProjectChain } from '@modules/dtos/common';
import { CreateProjectSchema } from '@modules/dtos/project';
import { useAppStore, useAppStoreSubscriber } from '@modules/stores/main/app-hooks';
import { Dialog, DialogContent, DialogTrigger } from '@modules/ui/dialog';
import { SidebarMenuButton } from '@modules/ui/sidebar';
import { cn } from '@modules/utils';
import { Plus } from 'lucide-react';
import * as React from 'react';
import { FormProvider, useForm, useFormContext } from 'react-hook-form';
import { CreateProjectInput, DialogBody } from './DialogBody';
import { DialogFooter } from './DialogFooter';
import { DialogHeader } from './DialogHeader';
import { ErrorBanner, ErrorBannerFormErrors, ErrorShape } from './ErrorBanner';

const useResetFormFromStore = () => {
	const proccesed = React.useRef(false);
	const ready = useAppStore((state) => state.ready);
	const draftProject = useAppStore((state) => state.draftProject);
	const draftProjectLogo = useAppStore((state) => state.draftProjectLogo);

	const ctx = useFormContext<CreateProjectInput>();

	React.useEffect(() => {
		if (!proccesed.current && ready) {
			proccesed.current = true;
			ctx.reset(draftProject);

			const uploaderId = draftProjectLogo?.pending_event.data.uploaderId;
			const projectId = draftProjectLogo?.key.split('/')[1] ?? '';
			if (uploaderId) {
				ctx.setValue('devWallet', uploaderId);
			}

			if (projectId) {
				ctx.setValue('id', projectId);
			}
		}

		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);
};

const RestoreDialog = () => {
	useResetFormFromStore();
	return <></>;
};

export function AddProjectDialog({ className }: { className?: string }) {
	const [open, setOpen] = React.useState(false);
	const [apiError, setApiError] = React.useState<ErrorShape | null>(null);

	const setDraftProject = useAppStore((state) => state.setDraftProject);

	const form = useForm<CreateProjectInput>({
		resolver: zodResolver(CreateProjectSchema),
		defaultValues: { chain: ProjectChain.SOLANA },
		mode: 'onChange',
	});

	useAppStoreSubscriber(
		(state) => state.wallet,
		(curr) => {
			if (curr?.walletAddress) {
				form.setValue('devWallet', curr.walletAddress);
			}
		},
	);

	useAppStoreSubscriber(
		(state) => state.draftProjectLogo,
		(curr) => {
			const uploaderId = curr?.pending_event.data.uploaderId;
			const projectId = curr?.key.split('/')[1] ?? '';
			if (uploaderId) {
				form.setValue('devWallet', uploaderId);
			}

			if (projectId) {
				form.setValue('id', projectId);
			}
		},
	);

	React.useEffect(() => {
		const sub = form.watch((values) => {
			setDraftProject(values);
		});
		return () => sub.unsubscribe();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<SidebarMenuButton
					variant="default"
					className={cn(
						'text-sm h-12 hover:cursor-pointer bg-primary font-semibold text-sidebar-primary-foreground',
						'hover:bg-sidebar-primary/90 hover:text-sidebar-primary-foreground/90 active:bg-sidebar-primary',
						'active:text-sidebar-primary-foreground',
						className,
						'group-data-[collapsible=icon]:items-center group-data-[collapsible=icon]:flex group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:!p-0',
					)}
				>
					<Plus className="!w-6 !h-6 mt-[2px] flex-shrink-0" />
					<span className="group-data-[collapsible=icon]:hidden">Shill your moon date &nbsp;&nbsp;🚀</span>
				</SidebarMenuButton>
			</DialogTrigger>

			<FormProvider {...form}>
				<DialogContent
					className="w-[90vw] sm:max-w-[720px] p-0 overflow-hidden"
					onOpenAutoFocus={(e) => e.preventDefault()}
					onCloseAutoFocus={(e) => e.preventDefault()}
				>
					<div className="max-h-[75vh] overflow-y-auto px-6 pt-4 pb-2">
						<RestoreDialog />
						<DialogHeader />

						<ErrorBanner data={apiError} />
						<ErrorBannerFormErrors />
						<DialogBody />
					</div>
					<div className="border-t px-6 py-3">
						<DialogFooter setApiError={setApiError} setOpen={setOpen} />
					</div>
				</DialogContent>
			</FormProvider>
		</Dialog>
	);
}
