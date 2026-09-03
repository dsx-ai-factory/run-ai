module.exports = {
  id: 'internal-filters',
  decorators: {
    oas3: {
      'filter-internal-enum-values': () => ({
        Schema: {
          leave(schema) {
            const internalValues = schema['x-enum-internal'];
            delete schema['x-enum-internal'];
            if (Array.isArray(internalValues) && Array.isArray(schema.enum)) {
              schema.enum = schema.enum.filter(v => !internalValues.includes(v));
            }
          }
        }
      })
    }
  }
};
