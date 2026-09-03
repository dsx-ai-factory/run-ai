(function () {
  "use strict";

  var CC_PROFILE = "cc";
  var CE_PROFILE = "ce";
  var CC_VENDOR_LOGO_PATH = "vendor-logo-on-light.svg";
  var CE_VENDOR_LOGO_PATH = "vendor-logo-on-light.svg";
  var PRODUCT_LOGO_PATH = "/distribution-assets/product-logo.svg";
  var BOOT_CONFIG_TIMEOUT_MS = 4500;
  var requestedProfile = "";

  function getVendorLogoUrl(profile) {
    var path =
      profile === CE_PROFILE ? CE_VENDOR_LOGO_PATH : CC_VENDOR_LOGO_PATH;
    return "/distribution-assets/" + profile + "/" + path;
  }

  function getFaviconUrl(profile) {
    return "/distribution-assets/" + profile + "/favicon.ico";
  }

  function setDistributionFavicon(profile) {
    var resolvedProfile = profile === CE_PROFILE ? CE_PROFILE : CC_PROFILE;
    var favicon = document.querySelector('link[rel~="icon"]');

    if (!favicon) {
      favicon = document.createElement("link");
      favicon.rel = "icon";
      document.head.appendChild(favicon);
    }

    favicon.onerror =
      resolvedProfile === CC_PROFILE
        ? null
        : function () {
            favicon.onerror = null;
            favicon.href = getFaviconUrl(CC_PROFILE);
          };
    favicon.href = getFaviconUrl(resolvedProfile);
  }

  function loadImage(source, fallbackSource, className) {
    return new Promise(function (resolve, reject) {
      var image = new Image();
      var didFallBack = false;

      image.className = className;
      image.alt = "";
      image.draggable = false;
      image.onload = function () {
        resolve(image);
      };
      image.onerror = function () {
        if (!didFallBack && fallbackSource && source !== fallbackSource) {
          didFallBack = true;
          image.src = fallbackSource;
          return;
        }
        image.onerror = null;
        reject(new Error("Distribution logo failed to load"));
      };
      image.src = source;
    });
  }

  function applyDistributionLoginLogo(profile) {
    var header = document.querySelector("header.login-pf-header");
    var resolvedProfile = profile === CE_PROFILE ? CE_PROFILE : CC_PROFILE;

    if (!header) {
      return;
    }

    requestedProfile = resolvedProfile;

    Promise.all([
      loadImage(
        getVendorLogoUrl(resolvedProfile),
        resolvedProfile === CE_PROFILE ? getVendorLogoUrl(CC_PROFILE) : "",
        "distribution-assets-vendor-logo",
      ),
      loadImage(PRODUCT_LOGO_PATH, "", "distribution-assets-product-logo"),
    ])
      .then(function (images) {
        if (requestedProfile !== resolvedProfile) {
          return;
        }

        var logo = document.createElement("span");
        var vendorViewport = document.createElement("span");
        var existingLogo = header.querySelector(
          ".distribution-assets-login-logo",
        );

        logo.className = "distribution-assets-login-logo";
        logo.setAttribute("role", "img");
        logo.setAttribute("aria-label", "NVIDIA Run:ai");
        images[0].className +=
          images[0].src.indexOf("/distribution-assets/" + CC_PROFILE + "/") >= 0
            ? " distribution-assets-vendor-logo-profile-cc"
            : " distribution-assets-vendor-logo-profile-ce";
        vendorViewport.className = "distribution-assets-vendor-logo-viewport";
        vendorViewport.appendChild(images[0]);
        logo.appendChild(vendorViewport);
        logo.appendChild(images[1]);

        if (existingLogo) {
          existingLogo.remove();
        }
        header.classList.remove("distribution-assets-profile-" + CC_PROFILE);
        header.classList.remove("distribution-assets-profile-" + CE_PROFILE);
        header.classList.add("distribution-assets-profile-" + resolvedProfile);
        header.style.backgroundImage = "none";
        header.appendChild(logo);
      })
      .catch(function () {
        // Keep the fixed-size logo slot empty when an image is unavailable.
      });
  }

  function applyDistributionAssetsProfile(profile) {
    var resolvedProfile = profile === CE_PROFILE ? CE_PROFILE : CC_PROFILE;

    setDistributionFavicon(resolvedProfile);
    if (requestedProfile !== resolvedProfile) {
      applyDistributionLoginLogo(resolvedProfile);
    }
  }

  function resolveDistributionAssetsProfile() {
    var didResolve = false;
    var timeoutId;

    function finish(profile) {
      if (didResolve) {
        return;
      }

      didResolve = true;
      clearTimeout(timeoutId);
      applyDistributionAssetsProfile(profile);
    }

    timeoutId = setTimeout(function () {
      finish(CC_PROFILE);
    }, BOOT_CONFIG_TIMEOUT_MS);

    fetch("/api/v1/boot-config", {
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
      },
    })
      .then(function (response) {
        if (!response.ok) {
          throw new Error("Boot config request failed");
        }

        return response.json();
      })
      .then(function (config) {
        var profile =
          config.distributionAssets &&
          config.distributionAssets.profile === CE_PROFILE
            ? CE_PROFILE
            : CC_PROFILE;
        finish(profile);
      })
      .catch(function () {
        finish(CC_PROFILE);
      });
  }

  function bootstrapDistributionAssets() {
    resolveDistributionAssetsProfile();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", bootstrapDistributionAssets, {
      once: true,
    });
  } else {
    bootstrapDistributionAssets();
  }
})();
