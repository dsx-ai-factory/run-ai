### Test Migration Guidelines

#### Tests That May Not Fit
1. **Tests requiring significant resources (CPU/GPU).**
    - Carefully evaluate these tests to prevent resource contention when running in parallel.

2. **Tests with cleanup logic that could impact other tests.**
    - For example, deleting shared resources might cause failures in other tests.

3. **Tests involving toggling Feature Flags (FF).**
    - Feature Flags that change the application behavior may lead to inconsistent outcomes in subsequent tests.

---

Ensure these guidelines are followed to maintain stability and reliability across test suites.