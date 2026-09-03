package com.runai.keycloak.extensions.users;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;
import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.models.UserModel;
import org.keycloak.models.UserProvider;
import org.keycloak.services.resources.admin.fgap.AdminPermissionEvaluator;
import org.keycloak.services.resources.admin.fgap.UserPermissionEvaluator;
import jakarta.ws.rs.ForbiddenException;
import jakarta.ws.rs.core.Response;

import java.util.*;
import java.util.stream.Stream;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

class CustomUserResourceTest {

    @Mock
    private KeycloakSession session;

    @Mock
    private RealmModel realmModel;

    @Mock
    private AdminPermissionEvaluator auth;

    @Mock
    private UserPermissionEvaluator usersEval;

    private CustomUserResource resourceProvider;

    @BeforeEach
    void setUp() {
        MockitoAnnotations.openMocks(this);
        // Keycloak's admin REST layer authenticates the request and supplies the
        // permission evaluator; here we stub it so the functional tests focus on the
        // listing logic. Authorization enforcement is covered by the dedicated tests.
        when(auth.users()).thenReturn(usersEval);
        resourceProvider = new CustomUserResource(session, realmModel, auth);
    }

    private UserModel createUser(String id, String username, String email, Map<String, List<String>> attributes) {
        UserModel user = mock(UserModel.class);
        when(user.getId()).thenReturn(id);
        when(user.getUsername()).thenReturn(username);
        when(user.getEmail()).thenReturn(email);
        when(user.getAttributes()).thenReturn(attributes);
        return user;
    }

    private void mockUserProvider(UserModel... users) {
        UserProvider userProvider = mock(UserProvider.class);
        when(session.users()).thenReturn(userProvider);
        when(userProvider.searchForUserStream(
                nullable(RealmModel.class),
                nullable(Map.class),
                nullable(Integer.class),
                nullable(Integer.class)
        )).thenReturn(Stream.of(users));
    }

    @Test
    void testGetUsers() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000001"),
                "runai.update.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));

        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(2, users.size());
    }

    @Test
    void testDateFilters() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000001"),
                "runai.update.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user3 = createUser("3", "user3", "user2@example.com", Map.of(
                "runai.creation.time", List.of("168623999999"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686239999999"),
                "runai.update.time", List.of("168623999999"),
                "runai.created.by", List.of("admin")
        ));

        mockUserProvider(user1, user2, user3);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, ">1686240000000", null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(2, users.size());
        assertEquals("1", users.get(0).get("id"));
    }

    @Test
    void testDateFilterEquals() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000001"),
                "runai.update.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));

        mockUserProvider(user1, user2);

        // Test case for EQUALS operator
        Response responseEquals = resourceProvider.getUsers(null, null, 10, 0, null, null, "=1686240000001", null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), responseEquals.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseEqualsBody = (Map<String, Object>) responseEquals.getEntity();
        List<Map<String, String>> usersEquals = (List<Map<String, String>>) responseEqualsBody.get("users");
        assertNotNull(usersEquals);
        assertEquals(2, usersEquals.size());
    }

    @Test
    void testDateFilterNotEquals() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686412860"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686412860"),
                "runai.update.time", List.of("1686412860"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686499261"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686499261"),
                "runai.update.time", List.of("1686499261"),
                "runai.created.by", List.of("admin")
        ));

        mockUserProvider(user1, user2);

        Response responseNotEquals = resourceProvider.getUsers(null, null, 10, 0, null, null, "=1686412860", null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), responseNotEquals.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseNotEqualsBody = (Map<String, Object>) responseNotEquals.getEntity();
        List<Map<String, String>> usersNotEquals = (List<Map<String, String>>) responseNotEqualsBody.get("users");
        assertNotNull(usersNotEquals);
        assertEquals(1, usersNotEquals.size());
        assertEquals("1", usersNotEquals.get(0).get("id"));
    }

    @Test
    void testSorting() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("fasle"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000001"),
                "runai.update.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));

        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, null, 10, 0, "creationTime", "asc", null, null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(2, users.size());
        assertEquals("2", users.get(0).get("id"));
        assertEquals("1", users.get(1).get("id"));
    }

    @Test
    void testPagination() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001")
        ));

        UserModel user3 = createUser("3", "user3", "user3@example.com", Map.of(
                "runai.creation.time", List.of("1686240000002")
        ));

        mockUserProvider(user1, user2, user3);

        Response response = resourceProvider.getUsers(null, null, null, 1, null, null, null, null, null, null,null, null, null);
        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(2, users.size());
        assertEquals("2", users.get(0).get("id"));
    }

    @Test
    void testInvalidDateFilterFormat() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));

        mockUserProvider(user1);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, "invalidDateFilter", null, null, null,null, null, null);
        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(0, users.size());
    }

    @Test
    void testFilterByUsername() {
        UserModel user1 = createUser("1", "user1user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));

        UserModel user2 = createUser("2", "user2user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));

        UserModel user3 = createUser("3", "user3user3", "user3@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));

        mockUserProvider(user1, user2, user3);

        Response response = resourceProvider.getUsers("user2", null, 10, 0, null, null, null, null, null, null,null, null, null);
        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(1, users.size());
        assertEquals("2", users.get(0).get("id"));
    }

    @Test
    void testInvalidSortingParameter() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001")
        ));

        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, null, 10, 0, "invalidField", "asc", null, null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(2, users.size());
        assertEquals("1", users.get(0).get("id")); // Check ordering here based on your actual logic
    }

    @Test
    void testPaginationEdgeCase() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001")
        ));

        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, null, 1, 2, null, null, null, null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertTrue(users.isEmpty()); // No users should be returned
    }

    @Test
    void testNoUsersFound() {
        mockUserProvider(); // No users

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertTrue(users.isEmpty()); // No users should be returned
    }

    @Test
    void testMultipleFilters() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("false")
        ));

        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, ">1686240000000", "runai.is_local:true", null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(0, users.size()); // No users should match both filters
    }

    @Test
    void testNullOrEmptyParameters() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));

        UserModel user2 = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001")
        ));

        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, "", 10, 0, null, "", null, null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(2, users.size()); // Should return all users
    }

    @Test
    void testRegularUsers() {
        UserModel notRegularUser = createUser("1", "user1", null, Map.of(
                "runai.creation.time", List.of("1686240000001")
        ));

        UserModel regularUser = createUser("2", "user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001")
        ));

        mockUserProvider(notRegularUser, regularUser);// No users

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null,null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");
        assertNotNull(users);
        assertEquals(1, users.size()); // No users should be returned
    }

    @Test
    void testAllUserProps() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin"),
                "GROUPS", new ArrayList<>(Arrays.asList("group1", "group2"))
        ));

        mockUserProvider(user1);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());

        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");

        assertNotNull(users);
        assertEquals(1, users.size());

        Map<String, String> userMap = users.get(0);

        assertEquals("1", userMap.get("id"));
        assertEquals("user1", userMap.get("username"));
        assertEquals("user1@example.com", userMap.get("email"));
        assertEquals("1686240000000", userMap.get("createdAt"));
        assertEquals("true", userMap.get("isLocal"));
        assertEquals("1686240000000", userMap.get("lastLogin"));
        assertEquals("1686240000000", userMap.get("updatedAt"));
        assertEquals("admin", userMap.get("createdBy"));
        assertEquals("group1,group2", userMap.get("groups")); // assuming the attribute is returned as a comma-separated string
    }

    @Test
    void testNoGroups() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin"),
                "GROUPS", List.of("")
        ));

        mockUserProvider(user1);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());

        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");

        assertNotNull(users);
        assertEquals(1, users.size());

        Map<String, String> userMap = users.get(0);
        assertEquals("", userMap.get("groups")); // assuming the attribute is returned as a comma-separated string
    }

    @Test
    void testOneGroup() {
        UserModel user1 = createUser("1", "user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin"),
                "GROUPS", List.of("test")
        ));

        mockUserProvider(user1);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());

        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");

        assertNotNull(users);
        assertEquals(1, users.size());

        Map<String, String> userMap = users.get(0);
        assertEquals("test", userMap.get("groups")); // assuming the attribute is returned as a comma-separated string
    }

    @Test
    void testSearchParameterUsername() {
        UserModel user1 = createUser("1", "johndoe", "john.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user2 = createUser("2", "janedoe", "jane.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000001"),
                "runai.update.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user3 = createUser("3", "testuser", "test.user@example.com", Map.of(
                "runai.creation.time", List.of("1686240000002"),
                "runai.is_local", List.of("false"),
                "runai.login.time", List.of("1686240000002"),
                "runai.update.time", List.of("1686240000002"),
                "runai.created.by", List.of("user")
        ));

        mockUserProvider(user1, user2, user3);

        // Testing with search parameter matching "john"
        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, "admin", "john", null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());

        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");

        assertNotNull(users);
        assertEquals(1, users.size());
        assertEquals("1", users.get(0).get("id"));
    }

    @Test
    void testSearchParameterCreatedBy() {
        UserModel user1 = createUser("1", "johndoe", "john.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000000"),
                "runai.update.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user2 = createUser("2", "janedoe", "jane.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000001"),
                "runai.update.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user3 = createUser("3", "testuser", "admintest.user@example.com", Map.of(
                "runai.creation.time", List.of("1686240000002"),
                "runai.is_local", List.of("false"),
                "runai.login.time", List.of("1686240000002"),
                "runai.update.time", List.of("1686240000002"),
                "runai.created.by", List.of("user")
        ));

        mockUserProvider(user1, user2, user3);

        // Testing with search parameter matching "jane"
        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "admin", null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());

        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");

        assertNotNull(users);
        assertEquals(3, users.size());
        assertEquals("3", users.get(0).get("id")); // Ensure only the matching user is returned// No users should match
    }

    @Test
    void testSearchParameterNotExist() {
        UserModel user2 = createUser("2", "janedoe", "jane.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.is_local", List.of("true"),
                "runai.login.time", List.of("1686240000001"),
                "runai.update.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));

        UserModel user3 = createUser("3", "testuser", "test.user@example.com", Map.of(
                "runai.creation.time", List.of("1686240000002"),
                "runai.is_local", List.of("false"),
                "runai.login.time", List.of("1686240000002"),
                "runai.update.time", List.of("1686240000002"),
                "runai.created.by", List.of("user")
        ));

        mockUserProvider(user2, user3);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "nonexistent", null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());

        @SuppressWarnings("unchecked")
        Map<String, Object> responseBody = (Map<String, Object>) response.getEntity();
        List<Map<String, String>> users = (List<Map<String, String>>) responseBody.get("users");


        assertNotNull(users);
        assertTrue(users.isEmpty()); // No users should match
    }

    @Test
    void testFilterByUsernameCaseInsensitive() {
        UserModel user1 = createUser("1", "user1user1", "user1@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));
        UserModel user2 = createUser("2", "user2user2", "user2@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000")
        ));
        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers("USER2", null, 10, 0, null, null, null, null, null, null, null, null, null);
        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>) ((Map<String, Object>) response.getEntity()).get("users");
        assertNotNull(users);
        assertEquals(1, users.size());
        assertEquals("2", users.get(0).get("id"));
    }

    @Test
    void testSearchParameterUsernameCaseInsensitive() {
        UserModel user1 = createUser("1", "johndoe", "john.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));
        UserModel user2 = createUser("2", "janedoe", "jane.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.created.by", List.of("admin")
        ));
        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "JOHN", null, null);
        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>) ((Map<String, Object>) response.getEntity()).get("users");
        assertNotNull(users);
        assertEquals(1, users.size());
        assertEquals("1", users.get(0).get("id"));
    }

    @Test
    void testSearchParameterCreatedByCaseInsensitive() {
        UserModel user1 = createUser("1", "johndoe", "john.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));
        UserModel user2 = createUser("2", "janedoe", "jane.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000001"),
                "runai.created.by", List.of("superuser")
        ));
        mockUserProvider(user1, user2);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, "ADMIN", null, null, null);
        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>) ((Map<String, Object>) response.getEntity()).get("users");
        assertNotNull(users);
        assertEquals(1, users.size());
        assertEquals("1", users.get(0).get("id"));
    }

    // -------------------------------------------------------------------------
    // Full name tests
    // -------------------------------------------------------------------------

    private UserModel createUserWithName(String id, String username, String email,
                                         String firstName, String lastName,
                                         Map<String, List<String>> attributes) {
        UserModel user = createUser(id, username, email, attributes);
        when(user.getFirstName()).thenReturn(firstName);
        when(user.getLastName()).thenReturn(lastName);
        return user;
    }

    @Test
    void testFirstNameAndLastNameReturnedInResponse() {
        UserModel user = createUserWithName("1", "zoe.hall", "zoe.hall@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000"),
                       "runai.is_local", List.of("true"),
                       "runai.login.time", List.of("1686240000000"),
                       "runai.update.time", List.of("1686240000000"),
                       "runai.created.by", List.of("admin")));
        mockUserProvider(user);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, null, null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getEntity();
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>) body.get("users");
        assertEquals(1, users.size());
        assertEquals("Zoe", users.get(0).get("firstName"));
        assertEquals("Hall", users.get(0).get("lastName"));
    }

    @Test
    void testFilterByFullNamePartialMatch() {
        UserModel zoeHall = createUserWithName("1", "zoe.hall", "zoe.hall@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000")));
        UserModel zoeBorovoy = createUserWithName("2", "zoe.borovoy", "zoe.borovoy@test.com", "Zoe", "Borovoy",
                Map.of("runai.creation.time", List.of("1686240000001")));
        UserModel alice = createUserWithName("3", "alice.smith", "alice.smith@test.com", "Alice", "Smith",
                Map.of("runai.creation.time", List.of("1686240000002")));
        mockUserProvider(zoeHall, zoeBorovoy, alice);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, "Zoe", null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(2, users.size());
        assertTrue(users.stream().allMatch(u -> u.get("firstName").equals("Zoe")));
    }

    @Test
    void testFilterByFullNameMultiWord() {
        UserModel zoeHall = createUserWithName("1", "zoe.hall", "zoe.hall@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000")));
        UserModel zoeBorovoy = createUserWithName("2", "zoe.borovoy", "zoe.borovoy@test.com", "Zoe", "Borovoy",
                Map.of("runai.creation.time", List.of("1686240000001")));
        mockUserProvider(zoeHall, zoeBorovoy);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, "Zoe B", null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(1, users.size());
        assertEquals("Borovoy", users.get(0).get("lastName"));
    }

    @Test
    void testFilterByFullNameCaseInsensitive() {
        UserModel user = createUserWithName("1", "zoe.hall", "zoe.hall@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000")));
        UserModel other = createUserWithName("2", "alice", "alice@test.com", "Alice", "Smith",
                Map.of("runai.creation.time", List.of("1686240000001")));
        mockUserProvider(user, other);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, "zoe hall", null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(1, users.size());
        assertEquals("1", users.get(0).get("id"));
    }

    @Test
    void testFilterByFullNameNoMatch() {
        UserModel user = createUserWithName("1", "zoe.hall", "zoe.hall@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000")));
        mockUserProvider(user);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, "Nonexistent", null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertTrue(users.isEmpty());
    }

    @Test
    void testSortByFullNameAscending() {
        UserModel zoe = createUserWithName("1", "zoe", "zoe@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000")));
        UserModel alice = createUserWithName("2", "alice", "alice@test.com", "Alice", "Smith",
                Map.of("runai.creation.time", List.of("1686240000001")));
        UserModel bob = createUserWithName("3", "bob", "bob@test.com", "Bob", "Jones",
                Map.of("runai.creation.time", List.of("1686240000002")));
        mockUserProvider(zoe, alice, bob);

        Response response = resourceProvider.getUsers(null, null, 10, 0, "fullName", "asc", null, null, null, null, null, null, null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(3, users.size());
        assertEquals("2", users.get(0).get("id")); // Alice Smith
        assertEquals("3", users.get(1).get("id")); // Bob Jones
        assertEquals("1", users.get(2).get("id")); // Zoe Hall
    }

    @Test
    void testSortByFullNameDescending() {
        UserModel zoe = createUserWithName("1", "zoe", "zoe@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000")));
        UserModel alice = createUserWithName("2", "alice", "alice@test.com", "Alice", "Smith",
                Map.of("runai.creation.time", List.of("1686240000001")));
        mockUserProvider(zoe, alice);

        Response response = resourceProvider.getUsers(null, null, 10, 0, "fullName", "desc", null, null, null, null, null, null, null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(2, users.size());
        assertEquals("1", users.get(0).get("id")); // Zoe Hall first (desc)
        assertEquals("2", users.get(1).get("id")); // Alice Smith second
    }

    @Test
    void testGlobalSearchMatchesFullName() {
        UserModel zoeBorovoy = createUserWithName("1", "zoe.borovoy", "zoe.borovoy@test.com", "Zoe", "Borovoy",
                Map.of("runai.creation.time", List.of("1686240000000"),
                       "runai.created.by", List.of("admin")));
        UserModel alice = createUserWithName("2", "alice.smith", "alice.smith@test.com", "Alice", "Smith",
                Map.of("runai.creation.time", List.of("1686240000001"),
                       "runai.created.by", List.of("admin")));
        mockUserProvider(zoeBorovoy, alice);

        // "Zoe B" should match full name "Zoe Borovoy"
        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "Zoe B", null, null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(1, users.size());
        assertEquals("1", users.get(0).get("id"));
    }

    @Test
    void testGlobalSearchCaseInsensitiveFullName() {
        UserModel user = createUserWithName("1", "zoe", "zoe@test.com", "Zoe", "Hall",
                Map.of("runai.creation.time", List.of("1686240000000"),
                       "runai.created.by", List.of("admin")));
        UserModel other = createUserWithName("2", "alice", "alice@test.com", "Alice", "Smith",
                Map.of("runai.creation.time", List.of("1686240000001"),
                       "runai.created.by", List.of("admin")));
        mockUserProvider(user, other);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "zoe hall", null, null);

        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(1, users.size());
        assertEquals("1", users.get(0).get("id"));
    }

    @Test
    void testFullNameNullFieldsHandledGracefully() {
        UserModel user = createUser("1", "noname", "noname@test.com",
                Map.of("runai.creation.time", List.of("1686240000000")));
        when(user.getFirstName()).thenReturn(null);
        when(user.getLastName()).thenReturn(null);
        mockUserProvider(user);

        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, "something", null);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertTrue(users.isEmpty());
    }

    // -------------------------------------------------------------------------
    // Admin authorization tests
    // -------------------------------------------------------------------------

    @Test
    void testGetUsersRequiresViewUsersPermission() {
        mockUserProvider();

        resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, null, null);

        // The endpoint must gate on the realm-management "view-users" permission.
        verify(usersEval).requireView();
    }

    @Test
    void testGetUsersRejectedWhenNotPermitted() {
        doThrow(new ForbiddenException("requires view-users")).when(usersEval).requireView();

        assertThrows(ForbiddenException.class, () ->
                resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, null, null, null));

        // No user data should be fetched when authorization fails.
        verify(session, never()).users();
    }

    // -------------------------------------------------------------------------
    // Minimal search tests
    // -------------------------------------------------------------------------

    @Test
    void testMinimalSearchExcludesCreatedBy() {
        UserModel user = createUser("1", "johndoe", "john.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));
        mockUserProvider(user);

        // "admin" matches createdBy only — minimalSearch=true must not return the user
        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "admin", null, true);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertTrue(users.isEmpty(), "minimal search must not match on createdBy");
    }

    @Test
    void testMinimalSearchMatchesUsername() {
        UserModel user = createUser("1", "johndoe", "john.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));
        mockUserProvider(user);

        // "john" matches username — minimalSearch=true should return the user
        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "john", null, true);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(1, users.size(), "minimal search should match on username");
        assertEquals("1", users.get(0).get("id"));
    }

    @Test
    void testFullSearchIncludesCreatedBy() {
        UserModel user = createUser("1", "johndoe", "john.doe@example.com", Map.of(
                "runai.creation.time", List.of("1686240000000"),
                "runai.created.by", List.of("admin")
        ));
        mockUserProvider(user);

        // "admin" matches createdBy — minimalSearch=false (full search) must include the user
        Response response = resourceProvider.getUsers(null, null, 10, 0, null, null, null, null, null, null, "admin", null, false);

        assertEquals(Response.Status.OK.getStatusCode(), response.getStatus());
        @SuppressWarnings("unchecked")
        List<Map<String, String>> users = (List<Map<String, String>>)
                ((Map<String, Object>) response.getEntity()).get("users");
        assertEquals(1, users.size(), "full search should match on createdBy");
        assertEquals("1", users.get(0).get("id"));
    }
}
