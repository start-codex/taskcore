<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { Dialog as SheetPrimitive } from "bits-ui";

	let { ...restProps }: SheetPrimitive.PortalProps = $props();
</script>

<SheetPrimitive.Portal {...restProps} />
