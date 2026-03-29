<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";

	let { open = $bindable(false), ...restProps }: DropdownMenuPrimitive.SubProps = $props();
</script>

<DropdownMenuPrimitive.Sub bind:open {...restProps} />
