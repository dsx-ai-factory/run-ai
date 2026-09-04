package com.runai.keycloak.extensions.users;

import org.keycloak.models.KeycloakSession;
import org.keycloak.models.RealmModel;
import org.keycloak.models.UserModel;
import org.keycloak.services.resources.admin.fgap.AdminPermissionEvaluator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import jakarta.ws.rs.GET;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;
import jakarta.ws.rs.QueryParam;
import jakarta.ws.rs.Path;

/**
 * JAX-RS sub-resource returned by {@link CustomUserResourceProvider}. By the time this
 * resource is invoked, Keycloak's admin REST layer has already performed standard admin
 * authentication and supplied the {@link AdminPermissionEvaluator}; this resource only
 * needs to enforce the specific permission required for the operation.
 */
public class CustomUserResource {

    private static final Logger logger = LoggerFactory.getLogger(CustomUserResource.class);
    private static final String GREATER_THAN = ">";
    private static final String LESS_THAN = "<";
    private static final String EQUALS = "=";

    private static final String GROUPS = "GROUPS";
    private static final String CREATED_BY_PARAM = "createdBy";
    private static final String CREATED_BY = "runai.created.by";
    private static final String CREATION_TIME_PARAM = "creationTime";
    private static final String CREATION_TIME = "runai.creation.time";
    private static final String LOGIN_TIME_PARAM = "lastLogin";
    private static final String LOGIN_TIME = "runai.login.time";
    private static final String UPDATE_TIME_PARAM = "lastUpdated";
    private static final String UPDATE_TIME = "runai.update.time";
    public static final String RUNAI_IS_LOCAL_PARAM = "type";
    public static final String RUNAI_IS_LOCAL = "runai.is_local";

    public static final String USERS = "users";
    public static final String COUNT = "count";

    public static final String ID_KEY = "id";
    public static final String USERNAME_KEY = "username";
    public static final String EMAIL_KEY = "email";
    public static final String CREATED_AT_KEY = "createdAt";
    public static final String IS_LOCAL_KEY = "isLocal";
    public static final String LAST_LOGIN_KEY = "lastLogin";
    public static final String UPDATED_AT_KEY = "updatedAt";
    public static final String CREATED_BY_KEY = "createdBy";
    public static final String GROUPS_KEY = "groups";
    public static final String FIRST_NAME_KEY = "firstName";
    public static final String LAST_NAME_KEY = "lastName";
    public static final String USERNAME = "username";
    public static final String Q = "q";
    public static final String MAX = "max";
    public static final String OFFSET = "offset";
    public static final String SORT_BY = "sortBy";
    public static final String SORT_ORDER = "sortOrder";
    public static final String CREATION_FILTER = "creationFilter";
    public static final String LOGIN_FILTER = "loginFilter";
    public static final String UPDATE_FILTER = "updateFilter";
    public static final String CREATEDBY_KEY = "createdBy";
    public static final String DESC = "desc";
    public static final String SEARCH = "search";
    public static final String FULL_NAME = "fullName";
    public static final String MINIMAL_SEARCH = "minimalSearch";


    private final KeycloakSession session;
    private final RealmModel realm;
    private final AdminPermissionEvaluator auth;

    private static final int DEFAULT_MAX = 1000;
    private static final int DEFAULT_OFFSET = 0;
    private static final String DEFAULT_SORT_BY = "username";
    private static final String DEFAULT_SORT_ORDER = "asc";

    private static final int DAY_IN_SECONDS = 86400;


    public CustomUserResource(KeycloakSession session, RealmModel realm, AdminPermissionEvaluator auth) {
        this.session = session;
        this.realm = realm;
        this.auth = auth;
    }

    @GET
    @Produces(MediaType.APPLICATION_JSON)
    @Path("/users")
    public Response getUsers(
            @QueryParam(USERNAME) String username,
            @QueryParam(Q) String q,
            @QueryParam(MAX) Integer max,
            @QueryParam(OFFSET) Integer offset,
            @QueryParam(SORT_BY) String sortBy,
            @QueryParam(SORT_ORDER) String sortOrder,
            @QueryParam(CREATION_FILTER) String creationFilter,
            @QueryParam(LOGIN_FILTER) String loginFilter,
            @QueryParam(UPDATE_FILTER) String updateFilter,
            @QueryParam(CREATED_BY_KEY) String createdBy,
            @QueryParam(SEARCH) String search,
            @QueryParam(FULL_NAME) String fullName,
            @QueryParam(MINIMAL_SEARCH) Boolean minimalSearch){

        // Require the realm-management "view-users" permission before exposing any user
        // data (throws ForbiddenException otherwise). Authentication itself was already
        // performed by Keycloak's admin REST layer before this resource was created.
        auth.users().requireView();

        // Use default values if query parameters are not provided
        max = Optional.ofNullable(max).orElse(DEFAULT_MAX);
        offset = Optional.ofNullable(offset).orElse(DEFAULT_OFFSET);
        sortBy = Optional.ofNullable(sortBy).orElse(DEFAULT_SORT_BY);
        sortOrder = Optional.ofNullable(sortOrder).orElse(DEFAULT_SORT_ORDER);

        logger.debug("Fetching users with filters: username={}, createdBy={}, q={}, max={}, offset={}, sortBy={}, sortOrder={}, creationFilter={}, loginFilter={}, updateFilter={}",
                username, createdBy, q, max, offset, sortBy, sortOrder, creationFilter, loginFilter, updateFilter);

        Map<String, String> filters = parseQuery(q);
        logger.debug("Parsed filters: {}", filters);

        // Fetch users with initial filtering at the database level for performance
        Stream<UserModel> userStream = session.users()
                .searchForUserStream(realm, filters, null, null)
                .filter(this::isRegularUser);

        // Apply search on searchable fields if provided
        if (search != null && !search.trim().isEmpty()) {
            String searchLower = search.trim().toLowerCase(Locale.ROOT);
            boolean minimal = Boolean.TRUE.equals(minimalSearch);
            userStream = userStream.filter(user -> {
                String fn = user.getFirstName() != null ? user.getFirstName() : "";
                String ln = user.getLastName() != null ? user.getLastName() : "";
                String fullNameLower = (fn + " " + ln).trim().toLowerCase(Locale.ROOT);
                String email = user.getEmail() != null ? user.getEmail() : "";
                String uname = user.getUsername() != null ? user.getUsername() : "";
                if (minimal) {
                    return uname.toLowerCase(Locale.ROOT).contains(searchLower)
                            || fullNameLower.contains(searchLower);
                }
                return email.toLowerCase(Locale.ROOT).contains(searchLower)
                        || uname.toLowerCase(Locale.ROOT).contains(searchLower)
                        || getAttribute(user.getAttributes(), CREATED_BY).toLowerCase(Locale.ROOT).contains(searchLower)
                        || fullNameLower.contains(searchLower);
            });
        }
        if (username != null && !username.trim().isEmpty()) {
            String usernameLower = username.trim().toLowerCase(Locale.ROOT);
            userStream = userStream.filter(user -> {
                String email = user.getEmail() != null ? user.getEmail() : "";
                String uname = user.getUsername() != null ? user.getUsername() : "";
                return email.toLowerCase(Locale.ROOT).contains(usernameLower)
                        || uname.toLowerCase(Locale.ROOT).contains(usernameLower);
            });
        }
        if (fullName != null && !fullName.trim().isEmpty()) {
            String fullNameLower = fullName.trim().toLowerCase(Locale.ROOT);
            userStream = userStream.filter(user -> {
                String fn = user.getFirstName() != null ? user.getFirstName() : "";
                String ln = user.getLastName() != null ? user.getLastName() : "";
                return (fn + " " + ln).trim().toLowerCase(Locale.ROOT).contains(fullNameLower);
            });
        }
        if (createdBy != null && !createdBy.trim().isEmpty()) {
            userStream = userStream.filter(user -> matchesAttribute(user.getAttributes(), CREATED_BY, createdBy));
        }
        if (creationFilter != null || loginFilter != null || updateFilter != null) {
            logger.debug("Applying date filters: creationFilter={}, loginFilter={}, updateFilter={}",
                    creationFilter, loginFilter, updateFilter);
            userStream = userStream.filter(user -> applyDateFilters(user.getAttributes(), creationFilter, loginFilter, updateFilter));
        }

        List<UserModel> allUsers = userStream.collect(Collectors.toList());
        logger.debug("Number of users after filtering: {}", allUsers.size());

        // Apply sorting
        Comparator<UserModel> comparator = getComparator(sortBy);
        if (DESC.equalsIgnoreCase(sortOrder)) {
            comparator = comparator.reversed();
        }
        allUsers.sort(comparator);

        // Apply the pagination after sorting
        List<Map<String, String>> users = allUsers.stream()
                .skip(offset)
                .limit(max)
                .map(this::convertUserToMap)
                .collect(Collectors.toList());
        // Create the response map with users and count
        Map<String, Object> responseMap = new HashMap<>();
        responseMap.put(USERS, users);
        responseMap.put(COUNT, allUsers.size());

        return Response.ok(responseMap).build();
    }

    private boolean applyDateFilters(Map<String, List<String>> attributes, String creationFilter, String loginFilter, String updateFilter) {
        if (creationFilter != null && !matchesDateWithOperator(attributes, CREATION_TIME, creationFilter)) {
            return false;
        }
        if (loginFilter != null && !matchesDateWithOperator(attributes, LOGIN_TIME, loginFilter)) {
            return false;
        }
        if (updateFilter != null && !matchesDateWithOperator(attributes, UPDATE_TIME, updateFilter)) {
            return false;
        }
        return true;
    }

    public boolean isSameDay(long attributeEpochInSeconds, long filterDateEpochInSeconds) {
        long startOfDayFilter = (filterDateEpochInSeconds / DAY_IN_SECONDS) * DAY_IN_SECONDS;
        long endOfDayFilter = startOfDayFilter + (DAY_IN_SECONDS - 1);
        return attributeEpochInSeconds >= startOfDayFilter && attributeEpochInSeconds <= endOfDayFilter;
    }

    private boolean matchesDateWithOperator(Map<String, List<String>> attributes, String key, String filter) {
        try {
            String operator = extractOperator(filter);
            String dateString = extractDateValue(filter);

            logger.debug("Applying date filter for key {}: operator {}, date {}", key, operator, dateString);

            long filterEpochDate = Long.parseLong(dateString);

            String attributeValue = getAttribute(attributes, key);
            if (attributeValue.equals("")) {
                logger.warn("Attribute {} not found in user attributes", key);
                return false; // Attribute not found
            }

            long attributeEpochInSeconds = Long.parseLong(attributeValue);

            boolean result;
            switch (operator) {
                case GREATER_THAN:
                    result = attributeEpochInSeconds >= filterEpochDate;
                    break;
                case LESS_THAN:
                    result = attributeEpochInSeconds <= filterEpochDate;
                    break;
                case EQUALS:
                    long filterDateEpochInSeconds = Long.parseLong(dateString);
                    result = isSameDay(attributeEpochInSeconds, filterDateEpochInSeconds);
                    break;
                default:
                    logger.warn("Unsupported operator {} for key {}", operator, key);
                    return false; // Unsupported operator
            }

            logger.debug("Result of applying filter for key {} with operator {} and date {}: {}", key, operator, filterEpochDate, result);
            return result;
        } catch (NumberFormatException e) {
            logger.error("Error parsing date or attribute for key {} with filter {}: {}", key, filter, e.getMessage());
            return false; // Error parsing date or attribute
        }
    }

    private static String extractDateValue(String filter) {
        return filter.substring(1);
    }

    private static String extractOperator(String filter) {
        return filter.substring(0, 1);
    }

    private Comparator<UserModel> getComparator(String sortBy) {
        return Comparator.comparing(user -> {
            switch (sortBy) {
                case CREATED_BY_PARAM:
                    return getAttribute(user.getAttributes(), CREATED_BY);
                case CREATION_TIME_PARAM:
                    return getAttribute(user.getAttributes(), CREATION_TIME);
                case LOGIN_TIME_PARAM:
                    return getAttribute(user.getAttributes(), LOGIN_TIME);
                case UPDATE_TIME_PARAM:
                    return getAttribute(user.getAttributes(), UPDATE_TIME);
                case RUNAI_IS_LOCAL_PARAM:
                    return getAttribute(user.getAttributes(), RUNAI_IS_LOCAL);
                case FULL_NAME:
                    String fn = user.getFirstName() != null ? user.getFirstName() : "";
                    String ln = user.getLastName() != null ? user.getLastName() : "";
                    return (fn + " " + ln).trim().toLowerCase(Locale.ROOT);
                default:
                    return user.getEmail();
            }
        }, Comparator.nullsLast(String::compareTo)); // Ensure null-safe comparison
    }

    private Map<String, String> parseQuery(String q) {
        Map<String, String> filters = new HashMap<>();
        if (q != null && !q.trim().isEmpty()) {
            String[] params = q.split(",");
            for (String param : params) {
                String[] pair = param.split(":");
                if (pair.length == 2) {
                    filters.put(pair[0].trim(), pair[1].trim());
                } else {
                    logger.warn("Invalid query parameter format: {}", param);
                }
            }
        }
        return filters;
    }

    private boolean matchesAttribute(Map<String, List<String>> attributes, String key, String value) {
        if (!attributes.containsKey(key)) {
            return false; // Attribute key not found
        }

        String valueLower = value.trim().toLowerCase(Locale.ROOT);
        return attributes.get(key).stream()
                .filter(Objects::nonNull)
                .anyMatch(attrValue -> attrValue.toLowerCase(Locale.ROOT).contains(valueLower));
    }

    private String getAttribute(Map<String, List<String>> attributes, String key) {
        return attributes.getOrDefault(key, List.of("")).get(0);
    }

    private String getAttributes(Map<String, List<String>> attributes, String key) {
        return String.join(",", attributes.getOrDefault(key, List.of()));
    }

    private Map<String, String> convertUserToMap(UserModel user) {
        Map<String, String> userMap = new HashMap<>();
        userMap.put(ID_KEY, user.getId());
        userMap.put(USERNAME_KEY, user.getUsername());
        userMap.put(EMAIL_KEY, user.getEmail());
        userMap.put(CREATED_AT_KEY, getAttribute(user.getAttributes(), CREATION_TIME));
        userMap.put(IS_LOCAL_KEY, getAttribute(user.getAttributes(), RUNAI_IS_LOCAL));
        userMap.put(LAST_LOGIN_KEY, getAttribute(user.getAttributes(), LOGIN_TIME));
        userMap.put(UPDATED_AT_KEY, getAttribute(user.getAttributes(), UPDATE_TIME));
        userMap.put(CREATEDBY_KEY, getAttribute(user.getAttributes(), CREATED_BY));
        userMap.put(GROUPS_KEY, getAttributes(user.getAttributes(), GROUPS));
        userMap.put(FIRST_NAME_KEY, user.getFirstName() != null ? user.getFirstName() : "");
        userMap.put(LAST_NAME_KEY, user.getLastName() != null ? user.getLastName() : "");
        return userMap;
    }

    private boolean isRegularUser(UserModel user) {
        if (user.getServiceAccountClientLink() != null) {
            return false;
        }

        if (user.getEmail() == null || user.getEmail().isEmpty()) {
            return false;
        }
        return true;
    }
}
