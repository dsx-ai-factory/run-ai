# Pinned public NGC fixtures

These public exact-tag responses make regression tests independent of live NGC changes. They are metadata fixtures, not model weights or runtime validation.

- `llama-3.1-8b-instruct-2.0.13.json`: existing exact-tag fixture; 24 GB NVFP4, 27 GB FP8 and 33 GB BF16 hardware-tier selection.
- `llama-3.1-70b-instruct-2.0.13.json`: retrieved 2026-10-28 from [NGC metadata API](https://api.ngc.nvidia.com/v2/nim/metadata?resourceId=nim%2Fmeta%2Fllama-3.1-70b-instruct&tag=2.0.13). FP8 TP2/4/8, PP1 minima are 57/31/18 GB per device. Profiles record H100 PCI `2330:10de` as tested.

Tests retain explicit TP/PP, precision/backend and per-device units. Do not replace a profile's minimum with a model-card value or derive another layout by dividing a TP1 floor.
