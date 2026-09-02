package com.runai.keycloak.broker.oidc.mappers;

import org.junit.jupiter.api.Test;
import org.keycloak.provider.ProviderConfigProperty;
import jakarta.ws.rs.WebApplicationException;
import jakarta.ws.rs.core.Response;

import java.util.*;

import static org.junit.jupiter.api.Assertions.*;

class ClaimValidatorTest {
    @Test
    void test_simple() throws Exception {
        String expected = "this is a test";
        Object value = "a";
        ClaimValidator.validateClaim(value, expected);
    }

    @Test
    void test_missing() throws Exception {
        String expected = "this is a test";
        Object value = "missing";
        WebApplicationException exception = assertThrows(WebApplicationException.class, () -> {
            ClaimValidator.validateClaim(value, expected);
        });
        assertEquals(Response.Status.UNAUTHORIZED.getStatusCode(), exception.getResponse().getStatus());
    }

    @Test
    void test_twowords() throws Exception {
        String expected = "this is a test";
        Object value = "this is";
        WebApplicationException exception = assertThrows(WebApplicationException.class, () -> {
            ClaimValidator.validateClaim(value, expected);
        });
        assertEquals(Response.Status.UNAUTHORIZED.getStatusCode(), exception.getResponse().getStatus());
    }

    @Test
    void test_null(){
        String expected = "this is a test";
        WebApplicationException exception = assertThrows(WebApplicationException.class, () -> {
            ClaimValidator.validateClaim(null, expected);
        });
        assertEquals(Response.Status.UNAUTHORIZED.getStatusCode(), exception.getResponse().getStatus());
    }

    @Test
    void test_emptyclaim(){
        String expected = "";
        WebApplicationException exception = assertThrows(WebApplicationException.class, () -> {
            ClaimValidator.validateClaim(null, expected);
        });
        assertEquals(Response.Status.UNAUTHORIZED.getStatusCode(), exception.getResponse().getStatus());
    }

    @Test
    void test_emptyclaim_null(){
        String expected = " ";
        WebApplicationException exception = assertThrows(WebApplicationException.class, () -> {
            ClaimValidator.validateClaim("test", expected);
        });
        assertEquals(Response.Status.UNAUTHORIZED.getStatusCode(), exception.getResponse().getStatus());
    }

    @Test
    void test_array(){
        String expected = "this is a test";
        Object value = List.of("this");
        WebApplicationException exception = assertThrows(WebApplicationException.class, () -> {
            ClaimValidator.validateClaim("", expected);
        });
        assertEquals(Response.Status.UNAUTHORIZED.getStatusCode(), exception.getResponse().getStatus());
    }

    @Test
    void testConfigProperties() {
        ClaimValidator mapper = new ClaimValidator();
        List<ProviderConfigProperty> properties = mapper.getConfigProperties();
        
        assertNotNull(properties);
        assertEquals(2, properties.size());
        
        // Check that our expected config properties are present
        boolean hasClaimProperty = properties.stream().anyMatch(p -> 
            ClaimValidator.CLAIM.equals(p.getName()) &&
            "Claim".equals(p.getLabel()) &&
            ProviderConfigProperty.STRING_TYPE.equals(p.getType()));
        
        boolean hasExpectedClaimValueProperty = properties.stream().anyMatch(p -> 
            ClaimValidator.EXPECTED_CLAIM_VALUE.equals(p.getName()) &&
            "Expected claim value".equals(p.getLabel()) &&
            ProviderConfigProperty.STRING_TYPE.equals(p.getType()));
        
        assertTrue(hasClaimProperty, "Should have CLAIM config property");
        assertTrue(hasExpectedClaimValueProperty, "Should have EXPECTED_CLAIM_VALUE config property");
    }
}