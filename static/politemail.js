$(function() {
    var tmplOption = Hogan.compile($('#tmpl-option').html());

    var addOption = function() {
        var cell = tmplOption.render();
        $('.option-grid .cell:last').before(cell);
    }

    addOption();

    $('#option-add').click(addOption);
});
