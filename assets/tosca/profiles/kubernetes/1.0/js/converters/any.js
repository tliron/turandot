
exports.convert = function(value) {
    for (let name in value) {
        let field = value[name];
        if (field !== undefined) {
            return field;
        }
    }
    return null;
};
