$(function() {
    var tmplOption = Hogan.compile($('#tmpl-option').html());

    $('#option-add').click(function() {
        var cell = tmplOption.render();
        $('.option-grid .cell:last').before(cell);
    });
});
