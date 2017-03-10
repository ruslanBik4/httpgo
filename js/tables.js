/**
 * Created by ruslan on 10.03.17.
 */
function GetCorrectValue( this_element ) {

    if ( this_element.value.search('checked') > -1 ) // для чекбоксов
        return this_element.value;

    switch ( this_element.type )
    {
        case 'select-one':
            return $( 'option[value=' + this_element.value + ' ]', this_element).text();
        case 'checkbox':
            return ( this_element.checked ? 'checked' : 'off' );
        default:
            return this_element.value.replace(/^\s+|\s+$/g, '');
    }
}
function SelectedTD() {
    var parent_tr = $(this).parent();
    flag = true;
    $('.selected').each( function() {
        if ( $('.td[name=' + this.name + ']', parent_tr).text().search( GetCorrectValue(this) ) == -1 )
        {
            flag=false;
            return flag;
        }
    });
    if (flag)
        parent_tr.show();
}
var sel_tr = '#table_body .tr';
function FilterIsModified( event, this_element ) {
    event = event || window.event;
    var thisvalue = this_element.value,
        regExist =  new RegExp( "(where|AND)\\s+'" + this_element.name + "'=[^&]+(?=(AND|&))?", 'i' );

    if ( (event.type == 'keyup') && (event.keyCode != 13) ) // ввод текста до Enter никак не обрабатываем
    {
        return true;
    }
    if ( (thisvalue.replace(/^\s+|\s+$/g, '') == '')
        || ( ( this_element.type == 'checkbox' ) && ( !this_element.checked ) ) ) // получили пустое значение - обнуляем фильтры
    {
        $(this_element).removeClass('selected');
        $( sel_tr + ':hidden .td[name=' + this_element.name + ']' ).each( SelectedTD );
        next_page = $('#aNextPage').attr('href');
        if (next_page) // докачаем с учетом фильтра
        {
            $('#aNextPage').attr( 'href', next_page.replace( regExist, '' ) ).click();
            $('#aSaveCSV').attr('href', $('#aSaveCSV').attr('href').replace( regExist, '' ) );
            $('#aSavePDF').attr('href', $('#aSavePDF').attr('href').replace( regExist, '' ) );

        }
        else
            $.get( document.href + ' #table_body', GetNewrecords );

    }
    else
    {
        thisvalue = GetCorrectValue( this_element  );
        $( 'input[type=image]', this_element.form ).show();

        if ( thisvalue.search('checked') > -1 ) // для чекбоксов
            otbor = $( sel_tr + ' .td[name=' + this_element.name + ']:has(input[' + thisvalue + '])' ); //подходящие элементы
        else
            otbor = $( sel_tr + ' .td[name=' + this_element.name + ']:contains(' + thisvalue + ')' ); //подходящие элементы

        if ( $(this_element).hasClass('selected') ) //
            otbor.each( SelectedTD ); // надо проверить походящие на соответствие другим фильтрам
// 		else
        $( sel_tr + ' .td[name=' + this_element.name + ']' ).not(otbor).parent().hide(); // скрываем неподходящие
// 		$('.selected').removeClass('selected');
        $(this_element).addClass('selected');
        next_page = $('#aNextPage').attr('href');
        if (next_page) // докачаем с учетом фильтра
        {
// 			if ( $( sel_tr + ':visible').length < 50 )
            next_page.replace( /offset=\d+/, 'offset=' + $( sel_tr + ':visible').length );

            $('#aNextPage').attr( 'href', AddNewFilter( next_page, this_element ) ).click();
        }
        // создаем ссылку для создания CSV
        if ( href = $('#aSaveCSV').attr('href') )
            $('#aSaveCSV').attr('href', AddNewFilter( href, this_element ) );
        if ( href = $('#aSavePDF').attr('href') )
            $('#aSavePDF').attr('href', AddNewFilter( href, this_element ) );

        FormIsModified( event, this_element.form ); // включаем кнопку сброса
// 	 }
    }

    return false;
}
// gолучаем условия для sql-запроса, по строкам дадим LIKE
function GetConditionFromElement(this_element) {
    switch ( this_element.type )
    {
        case 'select-one':
            return this_element.name  +  "=" + this_element.value; // из списка получаем ЖЕСТКОЕ равенство
        case 'checkbox':
            return this_element.name;
        default:
            return this_element.name  +  " REGEXP '" + this_element.value.replace(/^\s+|\s+$/g, '') + "'";
    }
}
// вставляем новое условие в фильтра
function AddNewFilter( href, this_element ) {
    var add_where = GetConditionFromElement(this_element),
        regExist  =  new RegExp( this_element.name + "=[^&]+(?=(AND|&))?", 'i' );

    if ( href.search( /table=/) > -1 )
        return href.replace( /table=([^&]+?(where[^&]+)*)(?=&)/i,
            function (str, p1, p2, offset, s)
            {
                if ( p1.search( this_element.name ) > -1 )
                    return "table=" + p1.replace( regExist, add_where );
                else
                    return "table=" + p1 + (p2 ? ' AND ' : ' where ' ) + add_where;
            } );
    if ( href.search( /where=/) > -1 )
        return href.replace( /where=([^&]+)/i,
            function (str, p1, offset, s)
            {
                if ( p1.search( this_element.name ) > -1 )
                    return str.replace( regExist, add_where );

                else
                    return str + ' AND ' +  add_where;

            } );
    else
        return href + "&where=" + add_where ;

}
// фильтрация
function PLayFilter( this_form ) {
    $('.selected').removeClass('selected').each( function () { this.value = ''; } );
    $( sel_tr ).show();
// 	 $(':input[form="' + this_form.name + '"][type!=image][modified]').each( function(i) { if (this.value) { cond = this.value; $('table tbody tr').find('td:eq(' + i + ')' ).each( function() { $.globalEval( 'cond1 = ($(this).text() ==' + cond + ');'); if (!cond1) $(this).parent().hide(); }) } });
    return false;
}