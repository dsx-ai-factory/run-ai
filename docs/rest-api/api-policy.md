# NVIDIA Run:ai API Policy

At NVIDIA Run:ai, we are dedicated to delivering stable, reliable, and well-documented APIs. Our goal is to ensure that our APIs evolve in a predictable, transparent manner, offering users a seamless experience.

NVIDIA Run:ai follows strict API design and operational standards to ensure a consistent and high-quality experience for users.

## Experimental APIs

Some APIs or API fields may be marked as Experimental. Experimental indicates that a feature is in early development and is provided for evaluation and feedback purposes. Experimental APIs:

* May be incomplete or unstable
* May change significantly or be removed in future versions
* Are not guaranteed to maintain backward compatibility

Use experimental APIs with caution, especially in production environments. Experimental APIs are clearly labeled as such in the documentation.

## API Life Cycle and Deprecation

While our goal is to maintain stable and backward-compatible APIs, there may be times when breaking changes or deprecations are necessary.

In case of breaking changes, the deprecated version of the API will be supported for two additional versions. During this period, no new features or functionality will be added to the deprecated API. When an API or API field is deprecated, the following process is followed:

* **Documentation** - The deprecated API or field is clearly labeled in the documentation, with a replacement provided where applicable.
* **Release Notes** - Information about deprecated APIs, including those scheduled for future removal, is included in the release notes.
* **Customer Notification** - Customers are notified of upcoming deprecations as part of the regular release communications.

## API Removal

After the defined backward compatibility period has ended, deprecated APIs or fields are removed from both the codebase and the documentation.
