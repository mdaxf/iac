/*
 * This combined file was created by the DataTables downloader builder:
 *   https://datatables.net/download
 *
 * To rebuild or modify this file with the latest versions of the included
 * software please visit:
 *   https://datatables.net/download/#ju/dt-1.13.4/e-2.1.3
 *
 * Included libraries:
 *   DataTables 1.13.4, Editor 2.1.3
 */

/*! DataTables 1.13.4
 * ©2008-2023 SpryMedia Ltd - datatables.net/license
 */

/**
 * @summary     DataTables
 * @description Paginate, search and order HTML tables
 * @version     1.13.4
 * @author      SpryMedia Ltd
 * @contact     www.datatables.net
 * @copyright   SpryMedia Ltd.
 *
 * This source file is free software, available under the following license:
 *   MIT license - http://datatables.net/license
 *
 * This source file is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY
 * or FITNESS FOR A PARTICULAR PURPOSE. See the license files for details.
 *
 * For details please refer to: http://www.datatables.net
 */

/*jslint evil: true, undef: true, browser: true */
/*globals $,require,jQuery,define,_selector_run,_selector_opts,_selector_first,_selector_row_indexes,_ext,_Api,_api_register,_api_registerPlural,_re_new_lines,_re_html,_re_formatted_numeric,_re_escape_regex,_empty,_intVal,_numToDecimal,_isNumber,_isHtml,_htmlNumeric,_pluck,_pluck_order,_range,_stripHtml,_unique,_fnBuildAjax,_fnAjaxUpdate,_fnAjaxParameters,_fnAjaxUpdateDraw,_fnAjaxDataSrc,_fnAddColumn,_fnColumnOptions,_fnAdjustColumnSizing,_fnVisibleToColumnIndex,_fnColumnIndexToVisible,_fnVisbleColumns,_fnGetColumns,_fnColumnTypes,_fnApplyColumnDefs,_fnHungarianMap,_fnCamelToHungarian,_fnLanguageCompat,_fnBrowserDetect,_fnAddData,_fnAddTr,_fnNodeToDataIndex,_fnNodeToColumnIndex,_fnGetCellData,_fnSetCellData,_fnSplitObjNotation,_fnGetObjectDataFn,_fnSetObjectDataFn,_fnGetDataMaster,_fnClearTable,_fnDeleteIndex,_fnInvalidate,_fnGetRowElements,_fnCreateTr,_fnBuildHead,_fnDrawHead,_fnDraw,_fnReDraw,_fnAddOptionsHtml,_fnDetectHeader,_fnGetUniqueThs,_fnFeatureHtmlFilter,_fnFilterComplete,_fnFilterCustom,_fnFilterColumn,_fnFilter,_fnFilterCreateSearch,_fnEscapeRegex,_fnFilterData,_fnFeatureHtmlInfo,_fnUpdateInfo,_fnInfoMacros,_fnInitialise,_fnInitComplete,_fnLengthChange,_fnFeatureHtmlLength,_fnFeatureHtmlPaginate,_fnPageChange,_fnFeatureHtmlProcessing,_fnProcessingDisplay,_fnFeatureHtmlTable,_fnScrollDraw,_fnApplyToChildren,_fnCalculateColumnWidths,_fnThrottle,_fnConvertToWidth,_fnGetWidestNode,_fnGetMaxLenString,_fnStringToCss,_fnSortFlatten,_fnSort,_fnSortAria,_fnSortListener,_fnSortAttachListener,_fnSortingClasses,_fnSortData,_fnSaveState,_fnLoadState,_fnSettingsFromNode,_fnLog,_fnMap,_fnBindAction,_fnCallbackReg,_fnCallbackFire,_fnLengthOverflow,_fnRenderer,_fnDataSource,_fnRowAttributes*/

(function( factory ) {
	"use strict";

	if ( typeof define === 'function' && define.amd ) {
		// AMD
		define( ['jquery'], function ( $ ) {
			return factory( $, window, document );
		} );
	}
	else if ( typeof exports === 'object' ) {
		// CommonJS
		// jQuery's factory checks for a global window - if it isn't present then it
		// returns a factory function that expects the window object
		var jq = require('jquery');

		if (typeof window !== 'undefined') {
			module.exports = function (root, $) {
				if ( ! root ) {
					// CommonJS environments without a window global must pass a
					// root. This will give an error otherwise
					root = window;
				}

				if ( ! $ ) {
					$ = jq( root );
				}

				return factory( $, root, root.document );
			};
		}
		else {
			return factory( jq, window, window.document );
		}
	}
	else {
		// Browser
		window.DataTable = factory( jQuery, window, document );
	}
}
(function( $, window, document, undefined ) {
	"use strict";

	
	var DataTable = function ( selector, options )
	{
		// Check if called with a window or jQuery object for DOM less applications
		// This is for backwards compatibility
		if (DataTable.factory(selector, options)) {
			return DataTable;
		}
	
		// When creating with `new`, create a new DataTable, returning the API instance
		if (this instanceof DataTable) {
			return $(selector).DataTable(options);
		}
		else {
			// Argument switching
			options = selector;
		}
	
		/**
		 * Perform a jQuery selector action on the table's TR elements (from the tbody) and
		 * return the resulting jQuery object.
		 *  @param {string|node|jQuery} sSelector jQuery selector or node collection to act on
		 *  @param {object} [oOpts] Optional parameters for modifying the rows to be included
		 *  @param {string} [oOpts.filter=none] Select TR elements that meet the current filter
		 *    criterion ("applied") or all TR elements (i.e. no filter).
		 *  @param {string} [oOpts.order=current] Order of the TR elements in the processed array.
		 *    Can be either 'current', whereby the current sorting of the table is used, or
		 *    'original' whereby the original order the data was read into the table is used.
		 *  @param {string} [oOpts.page=all] Limit the selection to the currently displayed page
		 *    ("current") or not ("all"). If 'current' is given, then order is assumed to be
		 *    'current' and filter is 'applied', regardless of what they might be given as.
		 *  @returns {object} jQuery object, filtered by the given selector.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Highlight every second row
		 *      oTable.$('tr:odd').css('backgroundColor', 'blue');
		 *    } );
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Filter to rows with 'Webkit' in them, add a background colour and then
		 *      // remove the filter, thus highlighting the 'Webkit' rows only.
		 *      oTable.fnFilter('Webkit');
		 *      oTable.$('tr', {"search": "applied"}).css('backgroundColor', 'blue');
		 *      oTable.fnFilter('');
		 *    } );
		 */
		this.$ = function ( sSelector, oOpts )
		{
			return this.api(true).$( sSelector, oOpts );
		};
		
		
		/**
		 * Almost identical to $ in operation, but in this case returns the data for the matched
		 * rows - as such, the jQuery selector used should match TR row nodes or TD/TH cell nodes
		 * rather than any descendants, so the data can be obtained for the row/cell. If matching
		 * rows are found, the data returned is the original data array/object that was used to
		 * create the row (or a generated array if from a DOM source).
		 *
		 * This method is often useful in-combination with $ where both functions are given the
		 * same parameters and the array indexes will match identically.
		 *  @param {string|node|jQuery} sSelector jQuery selector or node collection to act on
		 *  @param {object} [oOpts] Optional parameters for modifying the rows to be included
		 *  @param {string} [oOpts.filter=none] Select elements that meet the current filter
		 *    criterion ("applied") or all elements (i.e. no filter).
		 *  @param {string} [oOpts.order=current] Order of the data in the processed array.
		 *    Can be either 'current', whereby the current sorting of the table is used, or
		 *    'original' whereby the original order the data was read into the table is used.
		 *  @param {string} [oOpts.page=all] Limit the selection to the currently displayed page
		 *    ("current") or not ("all"). If 'current' is given, then order is assumed to be
		 *    'current' and filter is 'applied', regardless of what they might be given as.
		 *  @returns {array} Data for the matched elements. If any elements, as a result of the
		 *    selector, were not TR, TD or TH elements in the DataTable, they will have a null
		 *    entry in the array.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Get the data from the first row in the table
		 *      var data = oTable._('tr:first');
		 *
		 *      // Do something useful with the data
		 *      alert( "First cell is: "+data[0] );
		 *    } );
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Filter to 'Webkit' and get all data for
		 *      oTable.fnFilter('Webkit');
		 *      var data = oTable._('tr', {"search": "applied"});
		 *
		 *      // Do something with the data
		 *      alert( data.length+" rows matched the search" );
		 *    } );
		 */
		this._ = function ( sSelector, oOpts )
		{
			return this.api(true).rows( sSelector, oOpts ).data();
		};
		
		
		/**
		 * Create a DataTables Api instance, with the currently selected tables for
		 * the Api's context.
		 * @param {boolean} [traditional=false] Set the API instance's context to be
		 *   only the table referred to by the `DataTable.ext.iApiIndex` option, as was
		 *   used in the API presented by DataTables 1.9- (i.e. the traditional mode),
		 *   or if all tables captured in the jQuery object should be used.
		 * @return {DataTables.Api}
		 */
		this.api = function ( traditional )
		{
			return traditional ?
				new _Api(
					_fnSettingsFromNode( this[ _ext.iApiIndex ] )
				) :
				new _Api( this );
		};
		
		
		/**
		 * Add a single new row or multiple rows of data to the table. Please note
		 * that this is suitable for client-side processing only - if you are using
		 * server-side processing (i.e. "bServerSide": true), then to add data, you
		 * must add it to the data source, i.e. the server-side, through an Ajax call.
		 *  @param {array|object} data The data to be added to the table. This can be:
		 *    <ul>
		 *      <li>1D array of data - add a single row with the data provided</li>
		 *      <li>2D array of arrays - add multiple rows in a single call</li>
		 *      <li>object - data object when using <i>mData</i></li>
		 *      <li>array of objects - multiple data objects when using <i>mData</i></li>
		 *    </ul>
		 *  @param {bool} [redraw=true] redraw the table or not
		 *  @returns {array} An array of integers, representing the list of indexes in
		 *    <i>aoData</i> ({@link DataTable.models.oSettings}) that have been added to
		 *    the table.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    // Global var for counter
		 *    var giCount = 2;
		 *
		 *    $(document).ready(function() {
		 *      $('#example').dataTable();
		 *    } );
		 *
		 *    function fnClickAddRow() {
		 *      $('#example').dataTable().fnAddData( [
		 *        giCount+".1",
		 *        giCount+".2",
		 *        giCount+".3",
		 *        giCount+".4" ]
		 *      );
		 *
		 *      giCount++;
		 *    }
		 */
		this.fnAddData = function( data, redraw )
		{
			var api = this.api( true );
		
			/* Check if we want to add multiple rows or not */
			var rows = Array.isArray(data) && ( Array.isArray(data[0]) || $.isPlainObject(data[0]) ) ?
				api.rows.add( data ) :
				api.row.add( data );
		
			if ( redraw === undefined || redraw ) {
				api.draw();
			}
		
			return rows.flatten().toArray();
		};
		
		
		/**
		 * This function will make DataTables recalculate the column sizes, based on the data
		 * contained in the table and the sizes applied to the columns (in the DOM, CSS or
		 * through the sWidth parameter). This can be useful when the width of the table's
		 * parent element changes (for example a window resize).
		 *  @param {boolean} [bRedraw=true] Redraw the table or not, you will typically want to
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable( {
		 *        "sScrollY": "200px",
		 *        "bPaginate": false
		 *      } );
		 *
		 *      $(window).on('resize', function () {
		 *        oTable.fnAdjustColumnSizing();
		 *      } );
		 *    } );
		 */
		this.fnAdjustColumnSizing = function ( bRedraw )
		{
			var api = this.api( true ).columns.adjust();
			var settings = api.settings()[0];
			var scroll = settings.oScroll;
		
			if ( bRedraw === undefined || bRedraw ) {
				api.draw( false );
			}
			else if ( scroll.sX !== "" || scroll.sY !== "" ) {
				/* If not redrawing, but scrolling, we want to apply the new column sizes anyway */
				_fnScrollDraw( settings );
			}
		};
		
		
		/**
		 * Quickly and simply clear a table
		 *  @param {bool} [bRedraw=true] redraw the table or not
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Immediately 'nuke' the current rows (perhaps waiting for an Ajax callback...)
		 *      oTable.fnClearTable();
		 *    } );
		 */
		this.fnClearTable = function( bRedraw )
		{
			var api = this.api( true ).clear();
		
			if ( bRedraw === undefined || bRedraw ) {
				api.draw();
			}
		};
		
		
		/**
		 * The exact opposite of 'opening' a row, this function will close any rows which
		 * are currently 'open'.
		 *  @param {node} nTr the table row to 'close'
		 *  @returns {int} 0 on success, or 1 if failed (can't find the row)
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable;
		 *
		 *      // 'open' an information row when a row is clicked on
		 *      $('#example tbody tr').click( function () {
		 *        if ( oTable.fnIsOpen(this) ) {
		 *          oTable.fnClose( this );
		 *        } else {
		 *          oTable.fnOpen( this, "Temporary row opened", "info_row" );
		 *        }
		 *      } );
		 *
		 *      oTable = $('#example').dataTable();
		 *    } );
		 */
		this.fnClose = function( nTr )
		{
			this.api( true ).row( nTr ).child.hide();
		};
		
		
		/**
		 * Remove a row for the table
		 *  @param {mixed} target The index of the row from aoData to be deleted, or
		 *    the TR element you want to delete
		 *  @param {function|null} [callBack] Callback function
		 *  @param {bool} [redraw=true] Redraw the table or not
		 *  @returns {array} The row that was deleted
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Immediately remove the first row
		 *      oTable.fnDeleteRow( 0 );
		 *    } );
		 */
		this.fnDeleteRow = function( target, callback, redraw )
		{
			var api = this.api( true );
			var rows = api.rows( target );
			var settings = rows.settings()[0];
			var data = settings.aoData[ rows[0][0] ];
		
			rows.remove();
		
			if ( callback ) {
				callback.call( this, settings, data );
			}
		
			if ( redraw === undefined || redraw ) {
				api.draw();
			}
		
			return data;
		};
		
		
		/**
		 * Restore the table to it's original state in the DOM by removing all of DataTables
		 * enhancements, alterations to the DOM structure of the table and event listeners.
		 *  @param {boolean} [remove=false] Completely remove the table from the DOM
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      // This example is fairly pointless in reality, but shows how fnDestroy can be used
		 *      var oTable = $('#example').dataTable();
		 *      oTable.fnDestroy();
		 *    } );
		 */
		this.fnDestroy = function ( remove )
		{
			this.api( true ).destroy( remove );
		};
		
		
		/**
		 * Redraw the table
		 *  @param {bool} [complete=true] Re-filter and resort (if enabled) the table before the draw.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Re-draw the table - you wouldn't want to do it here, but it's an example :-)
		 *      oTable.fnDraw();
		 *    } );
		 */
		this.fnDraw = function( complete )
		{
			// Note that this isn't an exact match to the old call to _fnDraw - it takes
			// into account the new data, but can hold position.
			this.api( true ).draw( complete );
		};
		
		
		/**
		 * Filter the input based on data
		 *  @param {string} sInput String to filter the table on
		 *  @param {int|null} [iColumn] Column to limit filtering to
		 *  @param {bool} [bRegex=false] Treat as regular expression or not
		 *  @param {bool} [bSmart=true] Perform smart filtering or not
		 *  @param {bool} [bShowGlobal=true] Show the input global filter in it's input box(es)
		 *  @param {bool} [bCaseInsensitive=true] Do case-insensitive matching (true) or not (false)
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Sometime later - filter...
		 *      oTable.fnFilter( 'test string' );
		 *    } );
		 */
		this.fnFilter = function( sInput, iColumn, bRegex, bSmart, bShowGlobal, bCaseInsensitive )
		{
			var api = this.api( true );
		
			if ( iColumn === null || iColumn === undefined ) {
				api.search( sInput, bRegex, bSmart, bCaseInsensitive );
			}
			else {
				api.column( iColumn ).search( sInput, bRegex, bSmart, bCaseInsensitive );
			}
		
			api.draw();
		};
		
		
		/**
		 * Get the data for the whole table, an individual row or an individual cell based on the
		 * provided parameters.
		 *  @param {int|node} [src] A TR row node, TD/TH cell node or an integer. If given as
		 *    a TR node then the data source for the whole row will be returned. If given as a
		 *    TD/TH cell node then iCol will be automatically calculated and the data for the
		 *    cell returned. If given as an integer, then this is treated as the aoData internal
		 *    data index for the row (see fnGetPosition) and the data for that row used.
		 *  @param {int} [col] Optional column index that you want the data of.
		 *  @returns {array|object|string} If mRow is undefined, then the data for all rows is
		 *    returned. If mRow is defined, just data for that row, and is iCol is
		 *    defined, only data for the designated cell is returned.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    // Row data
		 *    $(document).ready(function() {
		 *      oTable = $('#example').dataTable();
		 *
		 *      oTable.$('tr').click( function () {
		 *        var data = oTable.fnGetData( this );
		 *        // ... do something with the array / object of data for the row
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Individual cell data
		 *    $(document).ready(function() {
		 *      oTable = $('#example').dataTable();
		 *
		 *      oTable.$('td').click( function () {
		 *        var sData = oTable.fnGetData( this );
		 *        alert( 'The cell clicked on had the value of '+sData );
		 *      } );
		 *    } );
		 */
		this.fnGetData = function( src, col )
		{
			var api = this.api( true );
		
			if ( src !== undefined ) {
				var type = src.nodeName ? src.nodeName.toLowerCase() : '';
		
				return col !== undefined || type == 'td' || type == 'th' ?
					api.cell( src, col ).data() :
					api.row( src ).data() || null;
			}
		
			return api.data().toArray();
		};
		
		
		/**
		 * Get an array of the TR nodes that are used in the table's body. Note that you will
		 * typically want to use the '$' API method in preference to this as it is more
		 * flexible.
		 *  @param {int} [iRow] Optional row index for the TR element you want
		 *  @returns {array|node} If iRow is undefined, returns an array of all TR elements
		 *    in the table's body, or iRow is defined, just the TR element requested.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Get the nodes from the table
		 *      var nNodes = oTable.fnGetNodes( );
		 *    } );
		 */
		this.fnGetNodes = function( iRow )
		{
			var api = this.api( true );
		
			return iRow !== undefined ?
				api.row( iRow ).node() :
				api.rows().nodes().flatten().toArray();
		};
		
		
		/**
		 * Get the array indexes of a particular cell from it's DOM element
		 * and column index including hidden columns
		 *  @param {node} node this can either be a TR, TD or TH in the table's body
		 *  @returns {int} If nNode is given as a TR, then a single index is returned, or
		 *    if given as a cell, an array of [row index, column index (visible),
		 *    column index (all)] is given.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      $('#example tbody td').click( function () {
		 *        // Get the position of the current data from the node
		 *        var aPos = oTable.fnGetPosition( this );
		 *
		 *        // Get the data array for this row
		 *        var aData = oTable.fnGetData( aPos[0] );
		 *
		 *        // Update the data array and return the value
		 *        aData[ aPos[1] ] = 'clicked';
		 *        this.innerHTML = 'clicked';
		 *      } );
		 *
		 *      // Init DataTables
		 *      oTable = $('#example').dataTable();
		 *    } );
		 */
		this.fnGetPosition = function( node )
		{
			var api = this.api( true );
			var nodeName = node.nodeName.toUpperCase();
		
			if ( nodeName == 'TR' ) {
				return api.row( node ).index();
			}
			else if ( nodeName == 'TD' || nodeName == 'TH' ) {
				var cell = api.cell( node ).index();
		
				return [
					cell.row,
					cell.columnVisible,
					cell.column
				];
			}
			return null;
		};
		
		
		/**
		 * Check to see if a row is 'open' or not.
		 *  @param {node} nTr the table row to check
		 *  @returns {boolean} true if the row is currently open, false otherwise
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable;
		 *
		 *      // 'open' an information row when a row is clicked on
		 *      $('#example tbody tr').click( function () {
		 *        if ( oTable.fnIsOpen(this) ) {
		 *          oTable.fnClose( this );
		 *        } else {
		 *          oTable.fnOpen( this, "Temporary row opened", "info_row" );
		 *        }
		 *      } );
		 *
		 *      oTable = $('#example').dataTable();
		 *    } );
		 */
		this.fnIsOpen = function( nTr )
		{
			return this.api( true ).row( nTr ).child.isShown();
		};
		
		
		/**
		 * This function will place a new row directly after a row which is currently
		 * on display on the page, with the HTML contents that is passed into the
		 * function. This can be used, for example, to ask for confirmation that a
		 * particular record should be deleted.
		 *  @param {node} nTr The table row to 'open'
		 *  @param {string|node|jQuery} mHtml The HTML to put into the row
		 *  @param {string} sClass Class to give the new TD cell
		 *  @returns {node} The row opened. Note that if the table row passed in as the
		 *    first parameter, is not found in the table, this method will silently
		 *    return.
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable;
		 *
		 *      // 'open' an information row when a row is clicked on
		 *      $('#example tbody tr').click( function () {
		 *        if ( oTable.fnIsOpen(this) ) {
		 *          oTable.fnClose( this );
		 *        } else {
		 *          oTable.fnOpen( this, "Temporary row opened", "info_row" );
		 *        }
		 *      } );
		 *
		 *      oTable = $('#example').dataTable();
		 *    } );
		 */
		this.fnOpen = function( nTr, mHtml, sClass )
		{
			return this.api( true )
				.row( nTr )
				.child( mHtml, sClass )
				.show()
				.child()[0];
		};
		
		
		/**
		 * Change the pagination - provides the internal logic for pagination in a simple API
		 * function. With this function you can have a DataTables table go to the next,
		 * previous, first or last pages.
		 *  @param {string|int} mAction Paging action to take: "first", "previous", "next" or "last"
		 *    or page number to jump to (integer), note that page 0 is the first page.
		 *  @param {bool} [bRedraw=true] Redraw the table or not
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *      oTable.fnPageChange( 'next' );
		 *    } );
		 */
		this.fnPageChange = function ( mAction, bRedraw )
		{
			var api = this.api( true ).page( mAction );
		
			if ( bRedraw === undefined || bRedraw ) {
				api.draw(false);
			}
		};
		
		
		/**
		 * Show a particular column
		 *  @param {int} iCol The column whose display should be changed
		 *  @param {bool} bShow Show (true) or hide (false) the column
		 *  @param {bool} [bRedraw=true] Redraw the table or not
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Hide the second column after initialisation
		 *      oTable.fnSetColumnVis( 1, false );
		 *    } );
		 */
		this.fnSetColumnVis = function ( iCol, bShow, bRedraw )
		{
			var api = this.api( true ).column( iCol ).visible( bShow );
		
			if ( bRedraw === undefined || bRedraw ) {
				api.columns.adjust().draw();
			}
		};
		
		
		/**
		 * Get the settings for a particular table for external manipulation
		 *  @returns {object} DataTables settings object. See
		 *    {@link DataTable.models.oSettings}
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *      var oSettings = oTable.fnSettings();
		 *
		 *      // Show an example parameter from the settings
		 *      alert( oSettings._iDisplayStart );
		 *    } );
		 */
		this.fnSettings = function()
		{
			return _fnSettingsFromNode( this[_ext.iApiIndex] );
		};
		
		
		/**
		 * Sort the table by a particular column
		 *  @param {int} iCol the data index to sort on. Note that this will not match the
		 *    'display index' if you have hidden data entries
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Sort immediately with columns 0 and 1
		 *      oTable.fnSort( [ [0,'asc'], [1,'asc'] ] );
		 *    } );
		 */
		this.fnSort = function( aaSort )
		{
			this.api( true ).order( aaSort ).draw();
		};
		
		
		/**
		 * Attach a sort listener to an element for a given column
		 *  @param {node} nNode the element to attach the sort listener to
		 *  @param {int} iColumn the column that a click on this node will sort on
		 *  @param {function} [fnCallback] callback function when sort is run
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *
		 *      // Sort on column 1, when 'sorter' is clicked on
		 *      oTable.fnSortListener( document.getElementById('sorter'), 1 );
		 *    } );
		 */
		this.fnSortListener = function( nNode, iColumn, fnCallback )
		{
			this.api( true ).order.listener( nNode, iColumn, fnCallback );
		};
		
		
		/**
		 * Update a table cell or row - this method will accept either a single value to
		 * update the cell with, an array of values with one element for each column or
		 * an object in the same format as the original data source. The function is
		 * self-referencing in order to make the multi column updates easier.
		 *  @param {object|array|string} mData Data to update the cell/row with
		 *  @param {node|int} mRow TR element you want to update or the aoData index
		 *  @param {int} [iColumn] The column to update, give as null or undefined to
		 *    update a whole row.
		 *  @param {bool} [bRedraw=true] Redraw the table or not
		 *  @param {bool} [bAction=true] Perform pre-draw actions or not
		 *  @returns {int} 0 on success, 1 on error
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *      oTable.fnUpdate( 'Example update', 0, 0 ); // Single cell
		 *      oTable.fnUpdate( ['a', 'b', 'c', 'd', 'e'], $('tbody tr')[0] ); // Row
		 *    } );
		 */
		this.fnUpdate = function( mData, mRow, iColumn, bRedraw, bAction )
		{
			var api = this.api( true );
		
			if ( iColumn === undefined || iColumn === null ) {
				api.row( mRow ).data( mData );
			}
			else {
				api.cell( mRow, iColumn ).data( mData );
			}
		
			if ( bAction === undefined || bAction ) {
				api.columns.adjust();
			}
		
			if ( bRedraw === undefined || bRedraw ) {
				api.draw();
			}
			return 0;
		};
		
		
		/**
		 * Provide a common method for plug-ins to check the version of DataTables being used, in order
		 * to ensure compatibility.
		 *  @param {string} sVersion Version string to check for, in the format "X.Y.Z". Note that the
		 *    formats "X" and "X.Y" are also acceptable.
		 *  @returns {boolean} true if this version of DataTables is greater or equal to the required
		 *    version, or false if this version of DataTales is not suitable
		 *  @method
		 *  @dtopt API
		 *  @deprecated Since v1.10
		 *
		 *  @example
		 *    $(document).ready(function() {
		 *      var oTable = $('#example').dataTable();
		 *      alert( oTable.fnVersionCheck( '1.9.0' ) );
		 *    } );
		 */
		this.fnVersionCheck = _ext.fnVersionCheck;
		
	
		var _that = this;
		var emptyInit = options === undefined;
		var len = this.length;
	
		if ( emptyInit ) {
			options = {};
		}
	
		this.oApi = this.internal = _ext.internal;
	
		// Extend with old style plug-in API methods
		for ( var fn in DataTable.ext.internal ) {
			if ( fn ) {
				this[fn] = _fnExternApiFunc(fn);
			}
		}
	
		this.each(function() {
			// For each initialisation we want to give it a clean initialisation
			// object that can be bashed around
			var o = {};
			var oInit = len > 1 ? // optimisation for single table case
				_fnExtend( o, options, true ) :
				options;
	
			/*global oInit,_that,emptyInit*/
			var i=0, iLen, j, jLen, k, kLen;
			var sId = this.getAttribute( 'id' );
			var bInitHandedOff = false;
			var defaults = DataTable.defaults;
			var $this = $(this);
			
			
			/* Sanity check */
			if ( this.nodeName.toLowerCase() != 'table' )
			{
				_fnLog( null, 0, 'Non-table node initialisation ('+this.nodeName+')', 2 );
				return;
			}
			
			/* Backwards compatibility for the defaults */
			_fnCompatOpts( defaults );
			_fnCompatCols( defaults.column );
			
			/* Convert the camel-case defaults to Hungarian */
			_fnCamelToHungarian( defaults, defaults, true );
			_fnCamelToHungarian( defaults.column, defaults.column, true );
			
			/* Setting up the initialisation object */
			_fnCamelToHungarian( defaults, $.extend( oInit, $this.data() ), true );
			
			
			
			/* Check to see if we are re-initialising a table */
			var allSettings = DataTable.settings;
			for ( i=0, iLen=allSettings.length ; i<iLen ; i++ )
			{
				var s = allSettings[i];
			
				/* Base check on table node */
				if (
					s.nTable == this ||
					(s.nTHead && s.nTHead.parentNode == this) ||
					(s.nTFoot && s.nTFoot.parentNode == this)
				) {
					var bRetrieve = oInit.bRetrieve !== undefined ? oInit.bRetrieve : defaults.bRetrieve;
					var bDestroy = oInit.bDestroy !== undefined ? oInit.bDestroy : defaults.bDestroy;
			
					if ( emptyInit || bRetrieve )
					{
						return s.oInstance;
					}
					else if ( bDestroy )
					{
						s.oInstance.fnDestroy();
						break;
					}
					else
					{
						_fnLog( s, 0, 'Cannot reinitialise DataTable', 3 );
						return;
					}
				}
			
				/* If the element we are initialising has the same ID as a table which was previously
				 * initialised, but the table nodes don't match (from before) then we destroy the old
				 * instance by simply deleting it. This is under the assumption that the table has been
				 * destroyed by other methods. Anyone using non-id selectors will need to do this manually
				 */
				if ( s.sTableId == this.id )
				{
					allSettings.splice( i, 1 );
					break;
				}
			}
			
			/* Ensure the table has an ID - required for accessibility */
			if ( sId === null || sId === "" )
			{
				sId = "DataTables_Table_"+(DataTable.ext._unique++);
				this.id = sId;
			}
			
			/* Create the settings object for this table and set some of the default parameters */
			var oSettings = $.extend( true, {}, DataTable.models.oSettings, {
				"sDestroyWidth": $this[0].style.width,
				"sInstance":     sId,
				"sTableId":      sId
			} );
			oSettings.nTable = this;
			oSettings.oApi   = _that.internal;
			oSettings.oInit  = oInit;
			
			allSettings.push( oSettings );
			
			// Need to add the instance after the instance after the settings object has been added
			// to the settings array, so we can self reference the table instance if more than one
			oSettings.oInstance = (_that.length===1) ? _that : $this.dataTable();
			
			// Backwards compatibility, before we apply all the defaults
			_fnCompatOpts( oInit );
			_fnLanguageCompat( oInit.oLanguage );
			
			// If the length menu is given, but the init display length is not, use the length menu
			if ( oInit.aLengthMenu && ! oInit.iDisplayLength )
			{
				oInit.iDisplayLength = Array.isArray( oInit.aLengthMenu[0] ) ?
					oInit.aLengthMenu[0][0] : oInit.aLengthMenu[0];
			}
			
			// Apply the defaults and init options to make a single init object will all
			// options defined from defaults and instance options.
			oInit = _fnExtend( $.extend( true, {}, defaults ), oInit );
			
			
			// Map the initialisation options onto the settings object
			_fnMap( oSettings.oFeatures, oInit, [
				"bPaginate",
				"bLengthChange",
				"bFilter",
				"bSort",
				"bSortMulti",
				"bInfo",
				"bProcessing",
				"bAutoWidth",
				"bSortClasses",
				"bServerSide",
				"bDeferRender"
			] );
			_fnMap( oSettings, oInit, [
				"asStripeClasses",
				"ajax",
				"fnServerData",
				"fnFormatNumber",
				"sServerMethod",
				"aaSorting",
				"aaSortingFixed",
				"aLengthMenu",
				"sPaginationType",
				"sAjaxSource",
				"sAjaxDataProp",
				"iStateDuration",
				"sDom",
				"bSortCellsTop",
				"iTabIndex",
				"fnStateLoadCallback",
				"fnStateSaveCallback",
				"renderer",
				"searchDelay",
				"rowId",
				[ "iCookieDuration", "iStateDuration" ], // backwards compat
				[ "oSearch", "oPreviousSearch" ],
				[ "aoSearchCols", "aoPreSearchCols" ],
				[ "iDisplayLength", "_iDisplayLength" ]
			] );
			_fnMap( oSettings.oScroll, oInit, [
				[ "sScrollX", "sX" ],
				[ "sScrollXInner", "sXInner" ],
				[ "sScrollY", "sY" ],
				[ "bScrollCollapse", "bCollapse" ]
			] );
			_fnMap( oSettings.oLanguage, oInit, "fnInfoCallback" );
			
			/* Callback functions which are array driven */
			_fnCallbackReg( oSettings, 'aoDrawCallback',       oInit.fnDrawCallback,      'user' );
			_fnCallbackReg( oSettings, 'aoServerParams',       oInit.fnServerParams,      'user' );
			_fnCallbackReg( oSettings, 'aoStateSaveParams',    oInit.fnStateSaveParams,   'user' );
			_fnCallbackReg( oSettings, 'aoStateLoadParams',    oInit.fnStateLoadParams,   'user' );
			_fnCallbackReg( oSettings, 'aoStateLoaded',        oInit.fnStateLoaded,       'user' );
			_fnCallbackReg( oSettings, 'aoRowCallback',        oInit.fnRowCallback,       'user' );
			_fnCallbackReg( oSettings, 'aoRowCreatedCallback', oInit.fnCreatedRow,        'user' );
			_fnCallbackReg( oSettings, 'aoHeaderCallback',     oInit.fnHeaderCallback,    'user' );
			_fnCallbackReg( oSettings, 'aoFooterCallback',     oInit.fnFooterCallback,    'user' );
			_fnCallbackReg( oSettings, 'aoInitComplete',       oInit.fnInitComplete,      'user' );
			_fnCallbackReg( oSettings, 'aoPreDrawCallback',    oInit.fnPreDrawCallback,   'user' );
			
			oSettings.rowIdFn = _fnGetObjectDataFn( oInit.rowId );
			
			/* Browser support detection */
			_fnBrowserDetect( oSettings );
			
			var oClasses = oSettings.oClasses;
			
			$.extend( oClasses, DataTable.ext.classes, oInit.oClasses );
			$this.addClass( oClasses.sTable );
			
			
			if ( oSettings.iInitDisplayStart === undefined )
			{
				/* Display start point, taking into account the save saving */
				oSettings.iInitDisplayStart = oInit.iDisplayStart;
				oSettings._iDisplayStart = oInit.iDisplayStart;
			}
			
			if ( oInit.iDeferLoading !== null )
			{
				oSettings.bDeferLoading = true;
				var tmp = Array.isArray( oInit.iDeferLoading );
				oSettings._iRecordsDisplay = tmp ? oInit.iDeferLoading[0] : oInit.iDeferLoading;
				oSettings._iRecordsTotal = tmp ? oInit.iDeferLoading[1] : oInit.iDeferLoading;
			}
			
			/* Language definitions */
			var oLanguage = oSettings.oLanguage;
			$.extend( true, oLanguage, oInit.oLanguage );
			
			if ( oLanguage.sUrl )
			{
				/* Get the language definitions from a file - because this Ajax call makes the language
				 * get async to the remainder of this function we use bInitHandedOff to indicate that
				 * _fnInitialise will be fired by the returned Ajax handler, rather than the constructor
				 */
				$.ajax( {
					dataType: 'json',
					url: oLanguage.sUrl,
					success: function ( json ) {
						_fnCamelToHungarian( defaults.oLanguage, json );
						_fnLanguageCompat( json );
						$.extend( true, oLanguage, json, oSettings.oInit.oLanguage );
			
						_fnCallbackFire( oSettings, null, 'i18n', [oSettings]);
						_fnInitialise( oSettings );
					},
					error: function () {
						// Error occurred loading language file, continue on as best we can
						_fnInitialise( oSettings );
					}
				} );
				bInitHandedOff = true;
			}
			else {
				_fnCallbackFire( oSettings, null, 'i18n', [oSettings]);
			}
			
			/*
			 * Stripes
			 */
			if ( oInit.asStripeClasses === null )
			{
				oSettings.asStripeClasses =[
					oClasses.sStripeOdd,
					oClasses.sStripeEven
				];
			}
			
			/* Remove row stripe classes if they are already on the table row */
			var stripeClasses = oSettings.asStripeClasses;
			var rowOne = $this.children('tbody').find('tr').eq(0);
			if ( $.inArray( true, $.map( stripeClasses, function(el, i) {
				return rowOne.hasClass(el);
			} ) ) !== -1 ) {
				$('tbody tr', this).removeClass( stripeClasses.join(' ') );
				oSettings.asDestroyStripes = stripeClasses.slice();
			}
			
			/*
			 * Columns
			 * See if we should load columns automatically or use defined ones
			 */
			var anThs = [];
			var aoColumnsInit;
			var nThead = this.getElementsByTagName('thead');
			if ( nThead.length !== 0 )
			{
				_fnDetectHeader( oSettings.aoHeader, nThead[0] );
				anThs = _fnGetUniqueThs( oSettings );
			}
			
			/* If not given a column array, generate one with nulls */
			if ( oInit.aoColumns === null )
			{
				aoColumnsInit = [];
				for ( i=0, iLen=anThs.length ; i<iLen ; i++ )
				{
					aoColumnsInit.push( null );
				}
			}
			else
			{
				aoColumnsInit = oInit.aoColumns;
			}
			
			/* Add the columns */
			for ( i=0, iLen=aoColumnsInit.length ; i<iLen ; i++ )
			{
				_fnAddColumn( oSettings, anThs ? anThs[i] : null );
			}
			
			/* Apply the column definitions */
			_fnApplyColumnDefs( oSettings, oInit.aoColumnDefs, aoColumnsInit, function (iCol, oDef) {
				_fnColumnOptions( oSettings, iCol, oDef );
			} );
			
			/* HTML5 attribute detection - build an mData object automatically if the
			 * attributes are found
			 */
			if ( rowOne.length ) {
				var a = function ( cell, name ) {
					return cell.getAttribute( 'data-'+name ) !== null ? name : null;
				};
			
				$( rowOne[0] ).children('th, td').each( function (i, cell) {
					var col = oSettings.aoColumns[i];
			
					if (! col) {
						_fnLog( oSettings, 0, 'Incorrect column count', 18 );
					}
					console.log(col)
					if ( col.mData === i ) {
						var sort = a( cell, 'sort' ) || a( cell, 'order' );
						var filter = a( cell, 'filter' ) || a( cell, 'search' );
			
						if ( sort !== null || filter !== null ) {
							col.mData = {
								_:      i+'.display',
								sort:   sort !== null   ? i+'.@data-'+sort   : undefined,
								type:   sort !== null   ? i+'.@data-'+sort   : undefined,
								filter: filter !== null ? i+'.@data-'+filter : undefined
							};
							col._isArrayHost = true;
			
							_fnColumnOptions( oSettings, i );
						}
					}
				} );
			}
			
			var features = oSettings.oFeatures;
			var loadedInit = function () {
				/*
				 * Sorting
				 * @todo For modularisation (1.11) this needs to do into a sort start up handler
				 */
			
				// If aaSorting is not defined, then we use the first indicator in asSorting
				// in case that has been altered, so the default sort reflects that option
				if ( oInit.aaSorting === undefined ) {
					var sorting = oSettings.aaSorting;
					for ( i=0, iLen=sorting.length ; i<iLen ; i++ ) {
						sorting[i][1] = oSettings.aoColumns[ i ].asSorting[0];
					}
				}
			
				/* Do a first pass on the sorting classes (allows any size changes to be taken into
				 * account, and also will apply sorting disabled classes if disabled
				 */
				_fnSortingClasses( oSettings );
			
				if ( features.bSort ) {
					_fnCallbackReg( oSettings, 'aoDrawCallback', function () {
						if ( oSettings.bSorted ) {
							var aSort = _fnSortFlatten( oSettings );
							var sortedColumns = {};
			
							$.each( aSort, function (i, val) {
								sortedColumns[ val.src ] = val.dir;
							} );
			
							_fnCallbackFire( oSettings, null, 'order', [oSettings, aSort, sortedColumns] );
							_fnSortAria( oSettings );
						}
					} );
				}
			
				_fnCallbackReg( oSettings, 'aoDrawCallback', function () {
					if ( oSettings.bSorted || _fnDataSource( oSettings ) === 'ssp' || features.bDeferRender ) {
						_fnSortingClasses( oSettings );
					}
				}, 'sc' );
			
			
				/*
				 * Final init
				 * Cache the header, body and footer as required, creating them if needed
				 */
			
				// Work around for Webkit bug 83867 - store the caption-side before removing from doc
				var captions = $this.children('caption').each( function () {
					this._captionSide = $(this).css('caption-side');
				} );
			
				var thead = $this.children('thead');
				if ( thead.length === 0 ) {
					thead = $('<thead/>').appendTo($this);
				}
				oSettings.nTHead = thead[0];
			
				var tbody = $this.children('tbody');
				if ( tbody.length === 0 ) {
					tbody = $('<tbody/>').insertAfter(thead);
				}
				oSettings.nTBody = tbody[0];
			
				var tfoot = $this.children('tfoot');
				if ( tfoot.length === 0 && captions.length > 0 && (oSettings.oScroll.sX !== "" || oSettings.oScroll.sY !== "") ) {
					// If we are a scrolling table, and no footer has been given, then we need to create
					// a tfoot element for the caption element to be appended to
					tfoot = $('<tfoot/>').appendTo($this);
				}
			
				if ( tfoot.length === 0 || tfoot.children().length === 0 ) {
					$this.addClass( oClasses.sNoFooter );
				}
				else if ( tfoot.length > 0 ) {
					oSettings.nTFoot = tfoot[0];
					_fnDetectHeader( oSettings.aoFooter, oSettings.nTFoot );
				}
			
				/* Check if there is data passing into the constructor */
				if ( oInit.aaData ) {
					for ( i=0 ; i<oInit.aaData.length ; i++ ) {
						_fnAddData( oSettings, oInit.aaData[ i ] );
					}
				}
				else if ( oSettings.bDeferLoading || _fnDataSource( oSettings ) == 'dom' ) {
					/* Grab the data from the page - only do this when deferred loading or no Ajax
					 * source since there is no point in reading the DOM data if we are then going
					 * to replace it with Ajax data
					 */
					_fnAddTr( oSettings, $(oSettings.nTBody).children('tr') );
				}
			
				/* Copy the data index array */
				oSettings.aiDisplay = oSettings.aiDisplayMaster.slice();
			
				/* Initialisation complete - table can be drawn */
				oSettings.bInitialised = true;
			
				/* Check if we need to initialise the table (it might not have been handed off to the
				 * language processor)
				 */
				if ( bInitHandedOff === false ) {
					_fnInitialise( oSettings );
				}
			};
			
			/* Must be done after everything which can be overridden by the state saving! */
			_fnCallbackReg( oSettings, 'aoDrawCallback', _fnSaveState, 'state_save' );
			
			if ( oInit.bStateSave )
			{
				features.bStateSave = true;
				_fnLoadState( oSettings, oInit, loadedInit );
			}
			else {
				loadedInit();
			}
			
		} );
		_that = null;
		return this;
	};
	
	
	/*
	 * It is useful to have variables which are scoped locally so only the
	 * DataTables functions can access them and they don't leak into global space.
	 * At the same time these functions are often useful over multiple files in the
	 * core and API, so we list, or at least document, all variables which are used
	 * by DataTables as private variables here. This also ensures that there is no
	 * clashing of variable names and that they can easily referenced for reuse.
	 */
	
	
	// Defined else where
	//  _selector_run
	//  _selector_opts
	//  _selector_first
	//  _selector_row_indexes
	
	var _ext; // DataTable.ext
	var _Api; // DataTable.Api
	var _api_register; // DataTable.Api.register
	var _api_registerPlural; // DataTable.Api.registerPlural
	
	var _re_dic = {};
	var _re_new_lines = /[\r\n\u2028]/g;
	var _re_html = /<.*?>/g;
	
	// This is not strict ISO8601 - Date.parse() is quite lax, although
	// implementations differ between browsers.
	var _re_date = /^\d{2,4}[\.\/\-]\d{1,2}[\.\/\-]\d{1,2}([T ]{1}\d{1,2}[:\.]\d{2}([\.:]\d{2})?)?$/;
	
	// Escape regular expression special characters
	var _re_escape_regex = new RegExp( '(\\' + [ '/', '.', '*', '+', '?', '|', '(', ')', '[', ']', '{', '}', '\\', '$', '^', '-' ].join('|\\') + ')', 'g' );
	
	// http://en.wikipedia.org/wiki/Foreign_exchange_market
	// - \u20BD - Russian ruble.
	// - \u20a9 - South Korean Won
	// - \u20BA - Turkish Lira
	// - \u20B9 - Indian Rupee
	// - R - Brazil (R$) and South Africa
	// - fr - Swiss Franc
	// - kr - Swedish krona, Norwegian krone and Danish krone
	// - \u2009 is thin space and \u202F is narrow no-break space, both used in many
	// - Ƀ - Bitcoin
	// - Ξ - Ethereum
	//   standards as thousands separators.
	var _re_formatted_numeric = /['\u00A0,$£€¥%\u2009\u202F\u20BD\u20a9\u20BArfkɃΞ]/gi;
	
	
	var _empty = function ( d ) {
		return !d || d === true || d === '-' ? true : false;
	};
	
	
	var _intVal = function ( s ) {
		var integer = parseInt( s, 10 );
		return !isNaN(integer) && isFinite(s) ? integer : null;
	};
	
	// Convert from a formatted number with characters other than `.` as the
	// decimal place, to a Javascript number
	var _numToDecimal = function ( num, decimalPoint ) {
		// Cache created regular expressions for speed as this function is called often
		if ( ! _re_dic[ decimalPoint ] ) {
			_re_dic[ decimalPoint ] = new RegExp( _fnEscapeRegex( decimalPoint ), 'g' );
		}
		return typeof num === 'string' && decimalPoint !== '.' ?
			num.replace( /\./g, '' ).replace( _re_dic[ decimalPoint ], '.' ) :
			num;
	};
	
	
	var _isNumber = function ( d, decimalPoint, formatted ) {
		let type = typeof d;
		var strType = type === 'string';
	
		if ( type === 'number' || type === 'bigint') {
			return true;
		}
	
		// If empty return immediately so there must be a number if it is a
		// formatted string (this stops the string "k", or "kr", etc being detected
		// as a formatted number for currency
		if ( _empty( d ) ) {
			return true;
		}
	
		if ( decimalPoint && strType ) {
			d = _numToDecimal( d, decimalPoint );
		}
	
		if ( formatted && strType ) {
			d = d.replace( _re_formatted_numeric, '' );
		}
	
		return !isNaN( parseFloat(d) ) && isFinite( d );
	};
	
	
	// A string without HTML in it can be considered to be HTML still
	var _isHtml = function ( d ) {
		return _empty( d ) || typeof d === 'string';
	};
	
	
	var _htmlNumeric = function ( d, decimalPoint, formatted ) {
		if ( _empty( d ) ) {
			return true;
		}
	
		var html = _isHtml( d );
		return ! html ?
			null :
			_isNumber( _stripHtml( d ), decimalPoint, formatted ) ?
				true :
				null;
	};
	
	
	var _pluck = function ( a, prop, prop2 ) {
		var out = [];
		var i=0, ien=a.length;
	
		// Could have the test in the loop for slightly smaller code, but speed
		// is essential here
		if ( prop2 !== undefined ) {
			for ( ; i<ien ; i++ ) {
				if ( a[i] && a[i][ prop ] ) {
					out.push( a[i][ prop ][ prop2 ] );
				}
			}
		}
		else {
			for ( ; i<ien ; i++ ) {
				if ( a[i] ) {
					out.push( a[i][ prop ] );
				}
			}
		}
	
		return out;
	};
	
	
	// Basically the same as _pluck, but rather than looping over `a` we use `order`
	// as the indexes to pick from `a`
	var _pluck_order = function ( a, order, prop, prop2 )
	{
		var out = [];
		var i=0, ien=order.length;
	
		// Could have the test in the loop for slightly smaller code, but speed
		// is essential here
		if ( prop2 !== undefined ) {
			for ( ; i<ien ; i++ ) {
				if ( a[ order[i] ][ prop ] ) {
					out.push( a[ order[i] ][ prop ][ prop2 ] );
				}
			}
		}
		else {
			for ( ; i<ien ; i++ ) {
				out.push( a[ order[i] ][ prop ] );
			}
		}
	
		return out;
	};
	
	
	var _range = function ( len, start )
	{
		var out = [];
		var end;
	
		if ( start === undefined ) {
			start = 0;
			end = len;
		}
		else {
			end = start;
			start = len;
		}
	
		for ( var i=start ; i<end ; i++ ) {
			out.push( i );
		}
	
		return out;
	};
	
	
	var _removeEmpty = function ( a )
	{
		var out = [];
	
		for ( var i=0, ien=a.length ; i<ien ; i++ ) {
			if ( a[i] ) { // careful - will remove all falsy values!
				out.push( a[i] );
			}
		}
	
		return out;
	};
	
	
	var _stripHtml = function ( d ) {
		return d.replace( _re_html, '' );
	};
	
	
	/**
	 * Determine if all values in the array are unique. This means we can short
	 * cut the _unique method at the cost of a single loop. A sorted array is used
	 * to easily check the values.
	 *
	 * @param  {array} src Source array
	 * @return {boolean} true if all unique, false otherwise
	 * @ignore
	 */
	var _areAllUnique = function ( src ) {
		if ( src.length < 2 ) {
			return true;
		}
	
		var sorted = src.slice().sort();
		var last = sorted[0];
	
		for ( var i=1, ien=sorted.length ; i<ien ; i++ ) {
			if ( sorted[i] === last ) {
				return false;
			}
	
			last = sorted[i];
		}
	
		return true;
	};
	
	
	/**
	 * Find the unique elements in a source array.
	 *
	 * @param  {array} src Source array
	 * @return {array} Array of unique items
	 * @ignore
	 */
	var _unique = function ( src )
	{
		if ( _areAllUnique( src ) ) {
			return src.slice();
		}
	
		// A faster unique method is to use object keys to identify used values,
		// but this doesn't work with arrays or objects, which we must also
		// consider. See jsperf.com/compare-array-unique-versions/4 for more
		// information.
		var
			out = [],
			val,
			i, ien=src.length,
			j, k=0;
	
		again: for ( i=0 ; i<ien ; i++ ) {
			val = src[i];
	
			for ( j=0 ; j<k ; j++ ) {
				if ( out[j] === val ) {
					continue again;
				}
			}
	
			out.push( val );
			k++;
		}
	
		return out;
	};
	
	// Surprisingly this is faster than [].concat.apply
	// https://jsperf.com/flatten-an-array-loop-vs-reduce/2
	var _flatten = function (out, val) {
		if (Array.isArray(val)) {
			for (var i=0 ; i<val.length ; i++) {
				_flatten(out, val[i]);
			}
		}
		else {
			out.push(val);
		}
	  
		return out;
	}
	
	var _includes = function (search, start) {
		if (start === undefined) {
			start = 0;
		}
	
		return this.indexOf(search, start) !== -1;	
	};
	
	// Array.isArray polyfill.
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/isArray
	if (! Array.isArray) {
	    Array.isArray = function(arg) {
	        return Object.prototype.toString.call(arg) === '[object Array]';
	    };
	}
	
	if (! Array.prototype.includes) {
		Array.prototype.includes = _includes;
	}
	
	// .trim() polyfill
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/trim
	if (!String.prototype.trim) {
	  String.prototype.trim = function () {
	    return this.replace(/^[\s\uFEFF\xA0]+|[\s\uFEFF\xA0]+$/g, '');
	  };
	}
	
	if (! String.prototype.includes) {
		String.prototype.includes = _includes;
	}
	
	/**
	 * DataTables utility methods
	 * 
	 * This namespace provides helper methods that DataTables uses internally to
	 * create a DataTable, but which are not exclusively used only for DataTables.
	 * These methods can be used by extension authors to save the duplication of
	 * code.
	 *
	 *  @namespace
	 */
	DataTable.util = {
		/**
		 * Throttle the calls to a function. Arguments and context are maintained
		 * for the throttled function.
		 *
		 * @param {function} fn Function to be called
		 * @param {integer} freq Call frequency in mS
		 * @return {function} Wrapped function
		 */
		throttle: function ( fn, freq ) {
			var
				frequency = freq !== undefined ? freq : 200,
				last,
				timer;
	
			return function () {
				var
					that = this,
					now  = +new Date(),
					args = arguments;
	
				if ( last && now < last + frequency ) {
					clearTimeout( timer );
	
					timer = setTimeout( function () {
						last = undefined;
						fn.apply( that, args );
					}, frequency );
				}
				else {
					last = now;
					fn.apply( that, args );
				}
			};
		},
	
	
		/**
		 * Escape a string such that it can be used in a regular expression
		 *
		 *  @param {string} val string to escape
		 *  @returns {string} escaped string
		 */
		escapeRegex: function ( val ) {
			return val.replace( _re_escape_regex, '\\$1' );
		},
	
		/**
		 * Create a function that will write to a nested object or array
		 * @param {*} source JSON notation string
		 * @returns Write function
		 */
		set: function ( source ) {
			if ( $.isPlainObject( source ) ) {
				/* Unlike get, only the underscore (global) option is used for for
				 * setting data since we don't know the type here. This is why an object
				 * option is not documented for `mData` (which is read/write), but it is
				 * for `mRender` which is read only.
				 */
				return DataTable.util.set( source._ );
			}
			else if ( source === null ) {
				// Nothing to do when the data source is null
				return function () {};
			}
			else if ( typeof source === 'function' ) {
				return function (data, val, meta) {
					source( data, 'set', val, meta );
				};
			}
			else if ( typeof source === 'string' && (source.indexOf('.') !== -1 ||
					  source.indexOf('[') !== -1 || source.indexOf('(') !== -1) )
			{
				// Like the get, we need to get data from a nested object
				var setData = function (data, val, src) {
					var a = _fnSplitObjNotation( src ), b;
					var aLast = a[a.length-1];
					var arrayNotation, funcNotation, o, innerSrc;
		
					for ( var i=0, iLen=a.length-1 ; i<iLen ; i++ ) {
						// Protect against prototype pollution
						if (a[i] === '__proto__' || a[i] === 'constructor') {
							throw new Error('Cannot set prototype values');
						}
		
						// Check if we are dealing with an array notation request
						arrayNotation = a[i].match(__reArray);
						funcNotation = a[i].match(__reFn);
		
						if ( arrayNotation ) {
							a[i] = a[i].replace(__reArray, '');
							data[ a[i] ] = [];
		
							// Get the remainder of the nested object to set so we can recurse
							b = a.slice();
							b.splice( 0, i+1 );
							innerSrc = b.join('.');
		
							// Traverse each entry in the array setting the properties requested
							if ( Array.isArray( val ) ) {
								for ( var j=0, jLen=val.length ; j<jLen ; j++ ) {
									o = {};
									setData( o, val[j], innerSrc );
									data[ a[i] ].push( o );
								}
							}
							else {
								// We've been asked to save data to an array, but it
								// isn't array data to be saved. Best that can be done
								// is to just save the value.
								data[ a[i] ] = val;
							}
		
							// The inner call to setData has already traversed through the remainder
							// of the source and has set the data, thus we can exit here
							return;
						}
						else if ( funcNotation ) {
							// Function call
							a[i] = a[i].replace(__reFn, '');
							data = data[ a[i] ]( val );
						}
		
						// If the nested object doesn't currently exist - since we are
						// trying to set the value - create it
						if ( data[ a[i] ] === null || data[ a[i] ] === undefined ) {
							data[ a[i] ] = {};
						}
						data = data[ a[i] ];
					}
		
					// Last item in the input - i.e, the actual set
					if ( aLast.match(__reFn ) ) {
						// Function call
						data = data[ aLast.replace(__reFn, '') ]( val );
					}
					else {
						// If array notation is used, we just want to strip it and use the property name
						// and assign the value. If it isn't used, then we get the result we want anyway
						data[ aLast.replace(__reArray, '') ] = val;
					}
				};
		
				return function (data, val) { // meta is also passed in, but not used
					return setData( data, val, source );
				};
			}
			else {
				// Array or flat object mapping
				return function (data, val) { // meta is also passed in, but not used
					data[source] = val;
				};
			}
		},
	
		/**
		 * Create a function that will read nested objects from arrays, based on JSON notation
		 * @param {*} source JSON notation string
		 * @returns Value read
		 */
		get: function ( source ) {
			if ( $.isPlainObject( source ) ) {
				// Build an object of get functions, and wrap them in a single call
				var o = {};
				$.each( source, function (key, val) {
					if ( val ) {
						o[key] = DataTable.util.get( val );
					}
				} );
		
				return function (data, type, row, meta) {
					var t = o[type] || o._;
					return t !== undefined ?
						t(data, type, row, meta) :
						data;
				};
			}
			else if ( source === null ) {
				// Give an empty string for rendering / sorting etc
				return function (data) { // type, row and meta also passed, but not used
					return data;
				};
			}
			else if ( typeof source === 'function' ) {
				return function (data, type, row, meta) {
					return source( data, type, row, meta );
				};
			}
			else if ( typeof source === 'string' && (source.indexOf('.') !== -1 ||
					  source.indexOf('[') !== -1 || source.indexOf('(') !== -1) )
			{
				/* If there is a . in the source string then the data source is in a
				 * nested object so we loop over the data for each level to get the next
				 * level down. On each loop we test for undefined, and if found immediately
				 * return. This allows entire objects to be missing and sDefaultContent to
				 * be used if defined, rather than throwing an error
				 */
				var fetchData = function (data, type, src) {
					var arrayNotation, funcNotation, out, innerSrc;
		
					if ( src !== "" ) {
						var a = _fnSplitObjNotation( src );
		
						for ( var i=0, iLen=a.length ; i<iLen ; i++ ) {
							// Check if we are dealing with special notation
							arrayNotation = a[i].match(__reArray);
							funcNotation = a[i].match(__reFn);
		
							if ( arrayNotation ) {
								// Array notation
								a[i] = a[i].replace(__reArray, '');
		
								// Condition allows simply [] to be passed in
								if ( a[i] !== "" ) {
									data = data[ a[i] ];
								}
								out = [];
		
								// Get the remainder of the nested object to get
								a.splice( 0, i+1 );
								innerSrc = a.join('.');
		
								// Traverse each entry in the array getting the properties requested
								if ( Array.isArray( data ) ) {
									for ( var j=0, jLen=data.length ; j<jLen ; j++ ) {
										out.push( fetchData( data[j], type, innerSrc ) );
									}
								}
		
								// If a string is given in between the array notation indicators, that
								// is used to join the strings together, otherwise an array is returned
								var join = arrayNotation[0].substring(1, arrayNotation[0].length-1);
								data = (join==="") ? out : out.join(join);
		
								// The inner call to fetchData has already traversed through the remainder
								// of the source requested, so we exit from the loop
								break;
							}
							else if ( funcNotation ) {
								// Function call
								a[i] = a[i].replace(__reFn, '');
								data = data[ a[i] ]();
								continue;
							}
		
							if ( data === null || data[ a[i] ] === undefined ) {
								return undefined;
							}
	
							data = data[ a[i] ];
						}
					}
		
					return data;
				};
		
				return function (data, type) { // row and meta also passed, but not used
					return fetchData( data, type, source );
				};
			}
			else {
				// Array or flat object mapping
				return function (data, type) { // row and meta also passed, but not used
					return data[source];
				};
			}
		}
	};
	
	
	
	/**
	 * Create a mapping object that allows camel case parameters to be looked up
	 * for their Hungarian counterparts. The mapping is stored in a private
	 * parameter called `_hungarianMap` which can be accessed on the source object.
	 *  @param {object} o
	 *  @memberof DataTable#oApi
	 */
	function _fnHungarianMap ( o )
	{
		var
			hungarian = 'a aa ai ao as b fn i m o s ',
			match,
			newKey,
			map = {};
	
		$.each( o, function (key, val) {
			match = key.match(/^([^A-Z]+?)([A-Z])/);
	
			if ( match && hungarian.indexOf(match[1]+' ') !== -1 )
			{
				newKey = key.replace( match[0], match[2].toLowerCase() );
				map[ newKey ] = key;
	
				if ( match[1] === 'o' )
				{
					_fnHungarianMap( o[key] );
				}
			}
		} );
	
		o._hungarianMap = map;
	}
	
	
	/**
	 * Convert from camel case parameters to Hungarian, based on a Hungarian map
	 * created by _fnHungarianMap.
	 *  @param {object} src The model object which holds all parameters that can be
	 *    mapped.
	 *  @param {object} user The object to convert from camel case to Hungarian.
	 *  @param {boolean} force When set to `true`, properties which already have a
	 *    Hungarian value in the `user` object will be overwritten. Otherwise they
	 *    won't be.
	 *  @memberof DataTable#oApi
	 */
	function _fnCamelToHungarian ( src, user, force )
	{
		if ( ! src._hungarianMap ) {
			_fnHungarianMap( src );
		}
	
		var hungarianKey;
	
		$.each( user, function (key, val) {
			hungarianKey = src._hungarianMap[ key ];
	
			if ( hungarianKey !== undefined && (force || user[hungarianKey] === undefined) )
			{
				// For objects, we need to buzz down into the object to copy parameters
				if ( hungarianKey.charAt(0) === 'o' )
				{
					// Copy the camelCase options over to the hungarian
					if ( ! user[ hungarianKey ] ) {
						user[ hungarianKey ] = {};
					}
					$.extend( true, user[hungarianKey], user[key] );
	
					_fnCamelToHungarian( src[hungarianKey], user[hungarianKey], force );
				}
				else {
					user[hungarianKey] = user[ key ];
				}
			}
		} );
	}
	
	
	/**
	 * Language compatibility - when certain options are given, and others aren't, we
	 * need to duplicate the values over, in order to provide backwards compatibility
	 * with older language files.
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnLanguageCompat( lang )
	{
		// Note the use of the Hungarian notation for the parameters in this method as
		// this is called after the mapping of camelCase to Hungarian
		var defaults = DataTable.defaults.oLanguage;
	
		// Default mapping
		var defaultDecimal = defaults.sDecimal;
		if ( defaultDecimal ) {
			_addNumericSort( defaultDecimal );
		}
	
		if ( lang ) {
			var zeroRecords = lang.sZeroRecords;
	
			// Backwards compatibility - if there is no sEmptyTable given, then use the same as
			// sZeroRecords - assuming that is given.
			if ( ! lang.sEmptyTable && zeroRecords &&
				defaults.sEmptyTable === "No data available in table" )
			{
				_fnMap( lang, lang, 'sZeroRecords', 'sEmptyTable' );
			}
	
			// Likewise with loading records
			if ( ! lang.sLoadingRecords && zeroRecords &&
				defaults.sLoadingRecords === "Loading..." )
			{
				_fnMap( lang, lang, 'sZeroRecords', 'sLoadingRecords' );
			}
	
			// Old parameter name of the thousands separator mapped onto the new
			if ( lang.sInfoThousands ) {
				lang.sThousands = lang.sInfoThousands;
			}
	
			var decimal = lang.sDecimal;
			if ( decimal && defaultDecimal !== decimal ) {
				_addNumericSort( decimal );
			}
		}
	}
	
	
	/**
	 * Map one parameter onto another
	 *  @param {object} o Object to map
	 *  @param {*} knew The new parameter name
	 *  @param {*} old The old parameter name
	 */
	var _fnCompatMap = function ( o, knew, old ) {
		if ( o[ knew ] !== undefined ) {
			o[ old ] = o[ knew ];
		}
	};
	
	
	/**
	 * Provide backwards compatibility for the main DT options. Note that the new
	 * options are mapped onto the old parameters, so this is an external interface
	 * change only.
	 *  @param {object} init Object to map
	 */
	function _fnCompatOpts ( init )
	{
		_fnCompatMap( init, 'ordering',      'bSort' );
		_fnCompatMap( init, 'orderMulti',    'bSortMulti' );
		_fnCompatMap( init, 'orderClasses',  'bSortClasses' );
		_fnCompatMap( init, 'orderCellsTop', 'bSortCellsTop' );
		_fnCompatMap( init, 'order',         'aaSorting' );
		_fnCompatMap( init, 'orderFixed',    'aaSortingFixed' );
		_fnCompatMap( init, 'paging',        'bPaginate' );
		_fnCompatMap( init, 'pagingType',    'sPaginationType' );
		_fnCompatMap( init, 'pageLength',    'iDisplayLength' );
		_fnCompatMap( init, 'searching',     'bFilter' );
	
		// Boolean initialisation of x-scrolling
		if ( typeof init.sScrollX === 'boolean' ) {
			init.sScrollX = init.sScrollX ? '100%' : '';
		}
		if ( typeof init.scrollX === 'boolean' ) {
			init.scrollX = init.scrollX ? '100%' : '';
		}
	
		// Column search objects are in an array, so it needs to be converted
		// element by element
		var searchCols = init.aoSearchCols;
	
		if ( searchCols ) {
			for ( var i=0, ien=searchCols.length ; i<ien ; i++ ) {
				if ( searchCols[i] ) {
					_fnCamelToHungarian( DataTable.models.oSearch, searchCols[i] );
				}
			}
		}
	}
	
	
	/**
	 * Provide backwards compatibility for column options. Note that the new options
	 * are mapped onto the old parameters, so this is an external interface change
	 * only.
	 *  @param {object} init Object to map
	 */
	function _fnCompatCols ( init )
	{
		_fnCompatMap( init, 'orderable',     'bSortable' );
		_fnCompatMap( init, 'orderData',     'aDataSort' );
		_fnCompatMap( init, 'orderSequence', 'asSorting' );
		_fnCompatMap( init, 'orderDataType', 'sortDataType' );
	
		// orderData can be given as an integer
		var dataSort = init.aDataSort;
		if ( typeof dataSort === 'number' && ! Array.isArray( dataSort ) ) {
			init.aDataSort = [ dataSort ];
		}
	}
	
	
	/**
	 * Browser feature detection for capabilities, quirks
	 *  @param {object} settings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnBrowserDetect( settings )
	{
		// We don't need to do this every time DataTables is constructed, the values
		// calculated are specific to the browser and OS configuration which we
		// don't expect to change between initialisations
		if ( ! DataTable.__browser ) {
			var browser = {};
			DataTable.__browser = browser;
	
			// Scrolling feature / quirks detection
			var n = $('<div/>')
				.css( {
					position: 'fixed',
					top: 0,
					left: $(window).scrollLeft()*-1, // allow for scrolling
					height: 1,
					width: 1,
					overflow: 'hidden'
				} )
				.append(
					$('<div/>')
						.css( {
							position: 'absolute',
							top: 1,
							left: 1,
							width: 100,
							overflow: 'scroll'
						} )
						.append(
							$('<div/>')
								.css( {
									width: '100%',
									height: 10
								} )
						)
				)
				.appendTo( 'body' );
	
			var outer = n.children();
			var inner = outer.children();
	
			// Numbers below, in order, are:
			// inner.offsetWidth, inner.clientWidth, outer.offsetWidth, outer.clientWidth
			//
			// IE6 XP:                           100 100 100  83
			// IE7 Vista:                        100 100 100  83
			// IE 8+ Windows:                     83  83 100  83
			// Evergreen Windows:                 83  83 100  83
			// Evergreen Mac with scrollbars:     85  85 100  85
			// Evergreen Mac without scrollbars: 100 100 100 100
	
			// Get scrollbar width
			browser.barWidth = outer[0].offsetWidth - outer[0].clientWidth;
	
			// IE6/7 will oversize a width 100% element inside a scrolling element, to
			// include the width of the scrollbar, while other browsers ensure the inner
			// element is contained without forcing scrolling
			browser.bScrollOversize = inner[0].offsetWidth === 100 && outer[0].clientWidth !== 100;
	
			// In rtl text layout, some browsers (most, but not all) will place the
			// scrollbar on the left, rather than the right.
			browser.bScrollbarLeft = Math.round( inner.offset().left ) !== 1;
	
			// IE8- don't provide height and width for getBoundingClientRect
			browser.bBounding = n[0].getBoundingClientRect().width ? true : false;
	
			n.remove();
		}
	
		$.extend( settings.oBrowser, DataTable.__browser );
		settings.oScroll.iBarWidth = DataTable.__browser.barWidth;
	}
	
	
	/**
	 * Array.prototype reduce[Right] method, used for browsers which don't support
	 * JS 1.6. Done this way to reduce code size, since we iterate either way
	 *  @param {object} settings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnReduce ( that, fn, init, start, end, inc )
	{
		var
			i = start,
			value,
			isSet = false;
	
		if ( init !== undefined ) {
			value = init;
			isSet = true;
		}
	
		while ( i !== end ) {
			if ( ! that.hasOwnProperty(i) ) {
				continue;
			}
	
			value = isSet ?
				fn( value, that[i], i, that ) :
				that[i];
	
			isSet = true;
			i += inc;
		}
	
		return value;
	}
	
	/**
	 * Add a column to the list used for the table with default values
	 *  @param {object} oSettings dataTables settings object
	 *  @param {node} nTh The th element for this column
	 *  @memberof DataTable#oApi
	 */
	function _fnAddColumn( oSettings, nTh )
	{
		// Add column to aoColumns array
		var oDefaults = DataTable.defaults.column;
		var iCol = oSettings.aoColumns.length;
		var oCol = $.extend( {}, DataTable.models.oColumn, oDefaults, {
			"nTh": nTh ? nTh : document.createElement('th'),
			"sTitle":    oDefaults.sTitle    ? oDefaults.sTitle    : nTh ? nTh.innerHTML : '',
			"aDataSort": oDefaults.aDataSort ? oDefaults.aDataSort : [iCol],
			"mData": oDefaults.mData ? oDefaults.mData : iCol,
			idx: iCol
		} );
		oSettings.aoColumns.push( oCol );
	
		// Add search object for column specific search. Note that the `searchCols[ iCol ]`
		// passed into extend can be undefined. This allows the user to give a default
		// with only some of the parameters defined, and also not give a default
		var searchCols = oSettings.aoPreSearchCols;
		searchCols[ iCol ] = $.extend( {}, DataTable.models.oSearch, searchCols[ iCol ] );
	
		// Use the default column options function to initialise classes etc
		_fnColumnOptions( oSettings, iCol, $(nTh).data() );
	}
	
	
	/**
	 * Apply options for a column
	 *  @param {object} oSettings dataTables settings object
	 *  @param {int} iCol column index to consider
	 *  @param {object} oOptions object with sType, bVisible and bSearchable etc
	 *  @memberof DataTable#oApi
	 */
	function _fnColumnOptions( oSettings, iCol, oOptions )
	{
		var oCol = oSettings.aoColumns[ iCol ];
		var oClasses = oSettings.oClasses;
		var th = $(oCol.nTh);
	
		// Try to get width information from the DOM. We can't get it from CSS
		// as we'd need to parse the CSS stylesheet. `width` option can override
		if ( ! oCol.sWidthOrig ) {
			// Width attribute
			oCol.sWidthOrig = th.attr('width') || null;
	
			// Style attribute
			var t = (th.attr('style') || '').match(/width:\s*(\d+[pxem%]+)/);
			if ( t ) {
				oCol.sWidthOrig = t[1];
			}
		}
	
		/* User specified column options */
		if ( oOptions !== undefined && oOptions !== null )
		{
			// Backwards compatibility
			_fnCompatCols( oOptions );
	
			// Map camel case parameters to their Hungarian counterparts
			_fnCamelToHungarian( DataTable.defaults.column, oOptions, true );
	
			/* Backwards compatibility for mDataProp */
			if ( oOptions.mDataProp !== undefined && !oOptions.mData )
			{
				oOptions.mData = oOptions.mDataProp;
			}
	
			if ( oOptions.sType )
			{
				oCol._sManualType = oOptions.sType;
			}
	
			// `class` is a reserved word in Javascript, so we need to provide
			// the ability to use a valid name for the camel case input
			if ( oOptions.className && ! oOptions.sClass )
			{
				oOptions.sClass = oOptions.className;
			}
			if ( oOptions.sClass ) {
				th.addClass( oOptions.sClass );
			}
	
			var origClass = oCol.sClass;
	
			$.extend( oCol, oOptions );
			_fnMap( oCol, oOptions, "sWidth", "sWidthOrig" );
	
			// Merge class from previously defined classes with this one, rather than just
			// overwriting it in the extend above
			if (origClass !== oCol.sClass) {
				oCol.sClass = origClass + ' ' + oCol.sClass;
			}
	
			/* iDataSort to be applied (backwards compatibility), but aDataSort will take
			 * priority if defined
			 */
			if ( oOptions.iDataSort !== undefined )
			{
				oCol.aDataSort = [ oOptions.iDataSort ];
			}
			_fnMap( oCol, oOptions, "aDataSort" );
		}
	
		/* Cache the data get and set functions for speed */
		var mDataSrc = oCol.mData;
		var mData = _fnGetObjectDataFn( mDataSrc );
		var mRender = oCol.mRender ? _fnGetObjectDataFn( oCol.mRender ) : null;
	
		var attrTest = function( src ) {
			return typeof src === 'string' && src.indexOf('@') !== -1;
		};
		oCol._bAttrSrc = $.isPlainObject( mDataSrc ) && (
			attrTest(mDataSrc.sort) || attrTest(mDataSrc.type) || attrTest(mDataSrc.filter)
		);
		oCol._setter = null;
	
		oCol.fnGetData = function (rowData, type, meta) {
			var innerData = mData( rowData, type, undefined, meta );
	
			return mRender && type ?
				mRender( innerData, type, rowData, meta ) :
				innerData;
		};
		oCol.fnSetData = function ( rowData, val, meta ) {
			return _fnSetObjectDataFn( mDataSrc )( rowData, val, meta );
		};
	
		// Indicate if DataTables should read DOM data as an object or array
		// Used in _fnGetRowElements
		if ( typeof mDataSrc !== 'number' && ! oCol._isArrayHost ) {
			oSettings._rowReadObject = true;
		}
	
		/* Feature sorting overrides column specific when off */
		if ( !oSettings.oFeatures.bSort )
		{
			oCol.bSortable = false;
			th.addClass( oClasses.sSortableNone ); // Have to add class here as order event isn't called
		}
	
		/* Check that the class assignment is correct for sorting */
		var bAsc = $.inArray('asc', oCol.asSorting) !== -1;
		var bDesc = $.inArray('desc', oCol.asSorting) !== -1;
		if ( !oCol.bSortable || (!bAsc && !bDesc) )
		{
			oCol.sSortingClass = oClasses.sSortableNone;
			oCol.sSortingClassJUI = "";
		}
		else if ( bAsc && !bDesc )
		{
			oCol.sSortingClass = oClasses.sSortableAsc;
			oCol.sSortingClassJUI = oClasses.sSortJUIAscAllowed;
		}
		else if ( !bAsc && bDesc )
		{
			oCol.sSortingClass = oClasses.sSortableDesc;
			oCol.sSortingClassJUI = oClasses.sSortJUIDescAllowed;
		}
		else
		{
			oCol.sSortingClass = oClasses.sSortable;
			oCol.sSortingClassJUI = oClasses.sSortJUI;
		}
	}
	
	
	/**
	 * Adjust the table column widths for new data. Note: you would probably want to
	 * do a redraw after calling this function!
	 *  @param {object} settings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnAdjustColumnSizing ( settings )
	{
		/* Not interested in doing column width calculation if auto-width is disabled */
		if ( settings.oFeatures.bAutoWidth !== false )
		{
			var columns = settings.aoColumns;
	
			_fnCalculateColumnWidths( settings );
			for ( var i=0 , iLen=columns.length ; i<iLen ; i++ )
			{
				columns[i].nTh.style.width = columns[i].sWidth;
			}
		}
	
		var scroll = settings.oScroll;
		if ( scroll.sY !== '' || scroll.sX !== '')
		{
			_fnScrollDraw( settings );
		}
	
		_fnCallbackFire( settings, null, 'column-sizing', [settings] );
	}
	
	
	/**
	 * Convert the index of a visible column to the index in the data array (take account
	 * of hidden columns)
	 *  @param {object} oSettings dataTables settings object
	 *  @param {int} iMatch Visible column index to lookup
	 *  @returns {int} i the data index
	 *  @memberof DataTable#oApi
	 */
	function _fnVisibleToColumnIndex( oSettings, iMatch )
	{
		var aiVis = _fnGetColumns( oSettings, 'bVisible' );
	
		return typeof aiVis[iMatch] === 'number' ?
			aiVis[iMatch] :
			null;
	}
	
	
	/**
	 * Convert the index of an index in the data array and convert it to the visible
	 *   column index (take account of hidden columns)
	 *  @param {int} iMatch Column index to lookup
	 *  @param {object} oSettings dataTables settings object
	 *  @returns {int} i the data index
	 *  @memberof DataTable#oApi
	 */
	function _fnColumnIndexToVisible( oSettings, iMatch )
	{
		var aiVis = _fnGetColumns( oSettings, 'bVisible' );
		var iPos = $.inArray( iMatch, aiVis );
	
		return iPos !== -1 ? iPos : null;
	}
	
	
	/**
	 * Get the number of visible columns
	 *  @param {object} oSettings dataTables settings object
	 *  @returns {int} i the number of visible columns
	 *  @memberof DataTable#oApi
	 */
	function _fnVisbleColumns( oSettings )
	{
		var vis = 0;
	
		// No reduce in IE8, use a loop for now
		$.each( oSettings.aoColumns, function ( i, col ) {
			if ( col.bVisible && $(col.nTh).css('display') !== 'none' ) {
				vis++;
			}
		} );
	
		return vis;
	}
	
	
	/**
	 * Get an array of column indexes that match a given property
	 *  @param {object} oSettings dataTables settings object
	 *  @param {string} sParam Parameter in aoColumns to look for - typically
	 *    bVisible or bSearchable
	 *  @returns {array} Array of indexes with matched properties
	 *  @memberof DataTable#oApi
	 */
	function _fnGetColumns( oSettings, sParam )
	{
		var a = [];
	
		$.map( oSettings.aoColumns, function(val, i) {
			if ( val[sParam] ) {
				a.push( i );
			}
		} );
	
		return a;
	}
	
	
	/**
	 * Calculate the 'type' of a column
	 *  @param {object} settings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnColumnTypes ( settings )
	{
		var columns = settings.aoColumns;
		var data = settings.aoData;
		var types = DataTable.ext.type.detect;
		var i, ien, j, jen, k, ken;
		var col, cell, detectedType, cache;
	
		// For each column, spin over the 
		for ( i=0, ien=columns.length ; i<ien ; i++ ) {
			col = columns[i];
			cache = [];
	
			if ( ! col.sType && col._sManualType ) {
				col.sType = col._sManualType;
			}
			else if ( ! col.sType ) {
				for ( j=0, jen=types.length ; j<jen ; j++ ) {
					for ( k=0, ken=data.length ; k<ken ; k++ ) {
						// Use a cache array so we only need to get the type data
						// from the formatter once (when using multiple detectors)
						if ( cache[k] === undefined ) {
							cache[k] = _fnGetCellData( settings, k, i, 'type' );
						}
	
						detectedType = types[j]( cache[k], settings );
	
						// If null, then this type can't apply to this column, so
						// rather than testing all cells, break out. There is an
						// exception for the last type which is `html`. We need to
						// scan all rows since it is possible to mix string and HTML
						// types
						if ( ! detectedType && j !== types.length-1 ) {
							break;
						}
	
						// Only a single match is needed for html type since it is
						// bottom of the pile and very similar to string - but it
						// must not be empty
						if ( detectedType === 'html' && ! _empty(cache[k]) ) {
							break;
						}
					}
	
					// Type is valid for all data points in the column - use this
					// type
					if ( detectedType ) {
						col.sType = detectedType;
						break;
					}
				}
	
				// Fall back - if no type was detected, always use string
				if ( ! col.sType ) {
					col.sType = 'string';
				}
			}
		}
	}
	
	
	/**
	 * Take the column definitions and static columns arrays and calculate how
	 * they relate to column indexes. The callback function will then apply the
	 * definition found for a column to a suitable configuration object.
	 *  @param {object} oSettings dataTables settings object
	 *  @param {array} aoColDefs The aoColumnDefs array that is to be applied
	 *  @param {array} aoCols The aoColumns array that defines columns individually
	 *  @param {function} fn Callback function - takes two parameters, the calculated
	 *    column index and the definition for that column.
	 *  @memberof DataTable#oApi
	 */
	function _fnApplyColumnDefs( oSettings, aoColDefs, aoCols, fn )
	{
		var i, iLen, j, jLen, k, kLen, def;
		var columns = oSettings.aoColumns;
	
		// Column definitions with aTargets
		if ( aoColDefs )
		{
			/* Loop over the definitions array - loop in reverse so first instance has priority */
			for ( i=aoColDefs.length-1 ; i>=0 ; i-- )
			{
				def = aoColDefs[i];
	
				/* Each definition can target multiple columns, as it is an array */
				var aTargets = def.target !== undefined
					? def.target
					: def.targets !== undefined
						? def.targets
						: def.aTargets;
	
				if ( ! Array.isArray( aTargets ) )
				{
					aTargets = [ aTargets ];
				}
	
				for ( j=0, jLen=aTargets.length ; j<jLen ; j++ )
				{
					if ( typeof aTargets[j] === 'number' && aTargets[j] >= 0 )
					{
						/* Add columns that we don't yet know about */
						while( columns.length <= aTargets[j] )
						{
							_fnAddColumn( oSettings );
						}
	
						/* Integer, basic index */
						fn( aTargets[j], def );
					}
					else if ( typeof aTargets[j] === 'number' && aTargets[j] < 0 )
					{
						/* Negative integer, right to left column counting */
						fn( columns.length+aTargets[j], def );
					}
					else if ( typeof aTargets[j] === 'string' )
					{
						/* Class name matching on TH element */
						for ( k=0, kLen=columns.length ; k<kLen ; k++ )
						{
							if ( aTargets[j] == "_all" ||
							     $(columns[k].nTh).hasClass( aTargets[j] ) )
							{
								fn( k, def );
							}
						}
					}
				}
			}
		}
	
		// Statically defined columns array
		if ( aoCols )
		{
			for ( i=0, iLen=aoCols.length ; i<iLen ; i++ )
			{
				fn( i, aoCols[i] );
			}
		}
	}
	
	/**
	 * Add a data array to the table, creating DOM node etc. This is the parallel to
	 * _fnGatherData, but for adding rows from a Javascript source, rather than a
	 * DOM source.
	 *  @param {object} oSettings dataTables settings object
	 *  @param {array} aData data array to be added
	 *  @param {node} [nTr] TR element to add to the table - optional. If not given,
	 *    DataTables will create a row automatically
	 *  @param {array} [anTds] Array of TD|TH elements for the row - must be given
	 *    if nTr is.
	 *  @returns {int} >=0 if successful (index of new aoData entry), -1 if failed
	 *  @memberof DataTable#oApi
	 */
	function _fnAddData ( oSettings, aDataIn, nTr, anTds )
	{
		/* Create the object for storing information about this new row */
		var iRow = oSettings.aoData.length;
		var oData = $.extend( true, {}, DataTable.models.oRow, {
			src: nTr ? 'dom' : 'data',
			idx: iRow
		} );
	
		oData._aData = aDataIn;
		oSettings.aoData.push( oData );
	
		/* Create the cells */
		var nTd, sThisType;
		var columns = oSettings.aoColumns;
	
		// Invalidate the column types as the new data needs to be revalidated
		for ( var i=0, iLen=columns.length ; i<iLen ; i++ )
		{
			columns[i].sType = null;
		}
	
		/* Add to the display array */
		oSettings.aiDisplayMaster.push( iRow );
	
		var id = oSettings.rowIdFn( aDataIn );
		if ( id !== undefined ) {
			oSettings.aIds[ id ] = oData;
		}
	
		/* Create the DOM information, or register it if already present */
		if ( nTr || ! oSettings.oFeatures.bDeferRender )
		{
			_fnCreateTr( oSettings, iRow, nTr, anTds );
		}
	
		return iRow;
	}
	
	
	/**
	 * Add one or more TR elements to the table. Generally we'd expect to
	 * use this for reading data from a DOM sourced table, but it could be
	 * used for an TR element. Note that if a TR is given, it is used (i.e.
	 * it is not cloned).
	 *  @param {object} settings dataTables settings object
	 *  @param {array|node|jQuery} trs The TR element(s) to add to the table
	 *  @returns {array} Array of indexes for the added rows
	 *  @memberof DataTable#oApi
	 */
	function _fnAddTr( settings, trs )
	{
		var row;
	
		// Allow an individual node to be passed in
		if ( ! (trs instanceof $) ) {
			trs = $(trs);
		}
	
		return trs.map( function (i, el) {
			row = _fnGetRowElements( settings, el );
			return _fnAddData( settings, row.data, el, row.cells );
		} );
	}
	
	
	/**
	 * Take a TR element and convert it to an index in aoData
	 *  @param {object} oSettings dataTables settings object
	 *  @param {node} n the TR element to find
	 *  @returns {int} index if the node is found, null if not
	 *  @memberof DataTable#oApi
	 */
	function _fnNodeToDataIndex( oSettings, n )
	{
		return (n._DT_RowIndex!==undefined) ? n._DT_RowIndex : null;
	}
	
	
	/**
	 * Take a TD element and convert it into a column data index (not the visible index)
	 *  @param {object} oSettings dataTables settings object
	 *  @param {int} iRow The row number the TD/TH can be found in
	 *  @param {node} n The TD/TH element to find
	 *  @returns {int} index if the node is found, -1 if not
	 *  @memberof DataTable#oApi
	 */
	function _fnNodeToColumnIndex( oSettings, iRow, n )
	{
		return $.inArray( n, oSettings.aoData[ iRow ].anCells );
	}
	
	
	/**
	 * Get the data for a given cell from the internal cache, taking into account data mapping
	 *  @param {object} settings dataTables settings object
	 *  @param {int} rowIdx aoData row id
	 *  @param {int} colIdx Column index
	 *  @param {string} type data get type ('display', 'type' 'filter|search' 'sort|order')
	 *  @returns {*} Cell data
	 *  @memberof DataTable#oApi
	 */
	function _fnGetCellData( settings, rowIdx, colIdx, type )
	{
		if (type === 'search') {
			type = 'filter';
		}
		else if (type === 'order') {
			type = 'sort';
		}
	
		var draw           = settings.iDraw;
		var col            = settings.aoColumns[colIdx];
		var rowData        = settings.aoData[rowIdx]._aData;
		var defaultContent = col.sDefaultContent;
		var cellData       = col.fnGetData( rowData, type, {
			settings: settings,
			row:      rowIdx,
			col:      colIdx
		} );
	
		if ( cellData === undefined ) {
			if ( settings.iDrawError != draw && defaultContent === null ) {
				_fnLog( settings, 0, "Requested unknown parameter "+
					(typeof col.mData=='function' ? '{function}' : "'"+col.mData+"'")+
					" for row "+rowIdx+", column "+colIdx, 4 );
				settings.iDrawError = draw;
			}
			return defaultContent;
		}
	
		// When the data source is null and a specific data type is requested (i.e.
		// not the original data), we can use default column data
		if ( (cellData === rowData || cellData === null) && defaultContent !== null && type !== undefined ) {
			cellData = defaultContent;
		}
		else if ( typeof cellData === 'function' ) {
			// If the data source is a function, then we run it and use the return,
			// executing in the scope of the data object (for instances)
			return cellData.call( rowData );
		}
	
		if ( cellData === null && type === 'display' ) {
			return '';
		}
	
		if ( type === 'filter' ) {
			var fomatters = DataTable.ext.type.search;
	
			if ( fomatters[ col.sType ] ) {
				cellData = fomatters[ col.sType ]( cellData );
			}
		}
	
		return cellData;
	}
	
	
	/**
	 * Set the value for a specific cell, into the internal data cache
	 *  @param {object} settings dataTables settings object
	 *  @param {int} rowIdx aoData row id
	 *  @param {int} colIdx Column index
	 *  @param {*} val Value to set
	 *  @memberof DataTable#oApi
	 */
	function _fnSetCellData( settings, rowIdx, colIdx, val )
	{
		var col     = settings.aoColumns[colIdx];
		var rowData = settings.aoData[rowIdx]._aData;
	
		col.fnSetData( rowData, val, {
			settings: settings,
			row:      rowIdx,
			col:      colIdx
		}  );
	}
	
	
	// Private variable that is used to match action syntax in the data property object
	var __reArray = /\[.*?\]$/;
	var __reFn = /\(\)$/;
	
	/**
	 * Split string on periods, taking into account escaped periods
	 * @param  {string} str String to split
	 * @return {array} Split string
	 */
	function _fnSplitObjNotation( str )
	{
		return $.map( str.match(/(\\.|[^\.])+/g) || [''], function ( s ) {
			return s.replace(/\\\./g, '.');
		} );
	}
	
	
	/**
	 * Return a function that can be used to get data from a source object, taking
	 * into account the ability to use nested objects as a source
	 *  @param {string|int|function} mSource The data source for the object
	 *  @returns {function} Data get function
	 *  @memberof DataTable#oApi
	 */
	var _fnGetObjectDataFn = DataTable.util.get;
	
	
	/**
	 * Return a function that can be used to set data from a source object, taking
	 * into account the ability to use nested objects as a source
	 *  @param {string|int|function} mSource The data source for the object
	 *  @returns {function} Data set function
	 *  @memberof DataTable#oApi
	 */
	var _fnSetObjectDataFn = DataTable.util.set;
	
	
	/**
	 * Return an array with the full table data
	 *  @param {object} oSettings dataTables settings object
	 *  @returns array {array} aData Master data array
	 *  @memberof DataTable#oApi
	 */
	function _fnGetDataMaster ( settings )
	{
		return _pluck( settings.aoData, '_aData' );
	}
	
	
	/**
	 * Nuke the table
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnClearTable( settings )
	{
		settings.aoData.length = 0;
		settings.aiDisplayMaster.length = 0;
		settings.aiDisplay.length = 0;
		settings.aIds = {};
	}
	
	
	 /**
	 * Take an array of integers (index array) and remove a target integer (value - not
	 * the key!)
	 *  @param {array} a Index array to target
	 *  @param {int} iTarget value to find
	 *  @memberof DataTable#oApi
	 */
	function _fnDeleteIndex( a, iTarget, splice )
	{
		var iTargetIndex = -1;
	
		for ( var i=0, iLen=a.length ; i<iLen ; i++ )
		{
			if ( a[i] == iTarget )
			{
				iTargetIndex = i;
			}
			else if ( a[i] > iTarget )
			{
				a[i]--;
			}
		}
	
		if ( iTargetIndex != -1 && splice === undefined )
		{
			a.splice( iTargetIndex, 1 );
		}
	}
	
	
	/**
	 * Mark cached data as invalid such that a re-read of the data will occur when
	 * the cached data is next requested. Also update from the data source object.
	 *
	 * @param {object} settings DataTables settings object
	 * @param {int}    rowIdx   Row index to invalidate
	 * @param {string} [src]    Source to invalidate from: undefined, 'auto', 'dom'
	 *     or 'data'
	 * @param {int}    [colIdx] Column index to invalidate. If undefined the whole
	 *     row will be invalidated
	 * @memberof DataTable#oApi
	 *
	 * @todo For the modularisation of v1.11 this will need to become a callback, so
	 *   the sort and filter methods can subscribe to it. That will required
	 *   initialisation options for sorting, which is why it is not already baked in
	 */
	function _fnInvalidate( settings, rowIdx, src, colIdx )
	{
		var row = settings.aoData[ rowIdx ];
		var i, ien;
		var cellWrite = function ( cell, col ) {
			// This is very frustrating, but in IE if you just write directly
			// to innerHTML, and elements that are overwritten are GC'ed,
			// even if there is a reference to them elsewhere
			while ( cell.childNodes.length ) {
				cell.removeChild( cell.firstChild );
			}
	
			cell.innerHTML = _fnGetCellData( settings, rowIdx, col, 'display' );
		};
	
		// Are we reading last data from DOM or the data object?
		if ( src === 'dom' || ((! src || src === 'auto') && row.src === 'dom') ) {
			// Read the data from the DOM
			row._aData = _fnGetRowElements(
					settings, row, colIdx, colIdx === undefined ? undefined : row._aData
				)
				.data;
		}
		else {
			// Reading from data object, update the DOM
			var cells = row.anCells;
	
			if ( cells ) {
				if ( colIdx !== undefined ) {
					cellWrite( cells[colIdx], colIdx );
				}
				else {
					for ( i=0, ien=cells.length ; i<ien ; i++ ) {
						cellWrite( cells[i], i );
					}
				}
			}
		}
	
		// For both row and cell invalidation, the cached data for sorting and
		// filtering is nulled out
		row._aSortData = null;
		row._aFilterData = null;
	
		// Invalidate the type for a specific column (if given) or all columns since
		// the data might have changed
		var cols = settings.aoColumns;
		if ( colIdx !== undefined ) {
			cols[ colIdx ].sType = null;
		}
		else {
			for ( i=0, ien=cols.length ; i<ien ; i++ ) {
				cols[i].sType = null;
			}
	
			// Update DataTables special `DT_*` attributes for the row
			_fnRowAttributes( settings, row );
		}
	}
	
	
	/**
	 * Build a data source object from an HTML row, reading the contents of the
	 * cells that are in the row.
	 *
	 * @param {object} settings DataTables settings object
	 * @param {node|object} TR element from which to read data or existing row
	 *   object from which to re-read the data from the cells
	 * @param {int} [colIdx] Optional column index
	 * @param {array|object} [d] Data source object. If `colIdx` is given then this
	 *   parameter should also be given and will be used to write the data into.
	 *   Only the column in question will be written
	 * @returns {object} Object with two parameters: `data` the data read, in
	 *   document order, and `cells` and array of nodes (they can be useful to the
	 *   caller, so rather than needing a second traversal to get them, just return
	 *   them from here).
	 * @memberof DataTable#oApi
	 */
	function _fnGetRowElements( settings, row, colIdx, d )
	{
		var
			tds = [],
			td = row.firstChild,
			name, col, o, i=0, contents,
			columns = settings.aoColumns,
			objectRead = settings._rowReadObject;
	
		// Allow the data object to be passed in, or construct
		d = d !== undefined ?
			d :
			objectRead ?
				{} :
				[];
	
		var attr = function ( str, td  ) {
			if ( typeof str === 'string' ) {
				var idx = str.indexOf('@');
	
				if ( idx !== -1 ) {
					var attr = str.substring( idx+1 );
					var setter = _fnSetObjectDataFn( str );
					setter( d, td.getAttribute( attr ) );
				}
			}
		};
	
		// Read data from a cell and store into the data object
		var cellProcess = function ( cell ) {
			if ( colIdx === undefined || colIdx === i ) {
				col = columns[i];
				contents = (cell.innerHTML).trim();
	
				if ( col && col._bAttrSrc ) {
					var setter = _fnSetObjectDataFn( col.mData._ );
					setter( d, contents );
	
					attr( col.mData.sort, cell );
					attr( col.mData.type, cell );
					attr( col.mData.filter, cell );
				}
				else {
					// Depending on the `data` option for the columns the data can
					// be read to either an object or an array.
					if ( objectRead ) {
						if ( ! col._setter ) {
							// Cache the setter function
							col._setter = _fnSetObjectDataFn( col.mData );
						}
						col._setter( d, contents );
					}
					else {
						d[i] = contents;
					}
				}
			}
	
			i++;
		};
	
		if ( td ) {
			// `tr` element was passed in
			while ( td ) {
				name = td.nodeName.toUpperCase();
	
				if ( name == "TD" || name == "TH" ) {
					cellProcess( td );
					tds.push( td );
				}
	
				td = td.nextSibling;
			}
		}
		else {
			// Existing row object passed in
			tds = row.anCells;
	
			for ( var j=0, jen=tds.length ; j<jen ; j++ ) {
				cellProcess( tds[j] );
			}
		}
	
		// Read the ID from the DOM if present
		var rowNode = row.firstChild ? row : row.nTr;
	
		if ( rowNode ) {
			var id = rowNode.getAttribute( 'id' );
	
			if ( id ) {
				_fnSetObjectDataFn( settings.rowId )( d, id );
			}
		}
	
		return {
			data: d,
			cells: tds
		};
	}
	/**
	 * Create a new TR element (and it's TD children) for a row
	 *  @param {object} oSettings dataTables settings object
	 *  @param {int} iRow Row to consider
	 *  @param {node} [nTrIn] TR element to add to the table - optional. If not given,
	 *    DataTables will create a row automatically
	 *  @param {array} [anTds] Array of TD|TH elements for the row - must be given
	 *    if nTr is.
	 *  @memberof DataTable#oApi
	 */
	function _fnCreateTr ( oSettings, iRow, nTrIn, anTds )
	{
		var
			row = oSettings.aoData[iRow],
			rowData = row._aData,
			cells = [],
			nTr, nTd, oCol,
			i, iLen, create;
	
		if ( row.nTr === null )
		{
			nTr = nTrIn || document.createElement('tr');
	
			row.nTr = nTr;
			row.anCells = cells;
	
			/* Use a private property on the node to allow reserve mapping from the node
			 * to the aoData array for fast look up
			 */
			nTr._DT_RowIndex = iRow;
	
			/* Special parameters can be given by the data source to be used on the row */
			_fnRowAttributes( oSettings, row );
	
			/* Process each column */
			for ( i=0, iLen=oSettings.aoColumns.length ; i<iLen ; i++ )
			{
				oCol = oSettings.aoColumns[i];
				create = nTrIn ? false : true;
	
				nTd = create ? document.createElement( oCol.sCellType ) : anTds[i];
	
				if (! nTd) {
					_fnLog( oSettings, 0, 'Incorrect column count', 18 );
				}
	
				nTd._DT_CellIndex = {
					row: iRow,
					column: i
				};
				
				cells.push( nTd );
	
				// Need to create the HTML if new, or if a rendering function is defined
				if ( create || ((oCol.mRender || oCol.mData !== i) &&
					 (!$.isPlainObject(oCol.mData) || oCol.mData._ !== i+'.display')
				)) {
					nTd.innerHTML = _fnGetCellData( oSettings, iRow, i, 'display' );
				}
	
				/* Add user defined class */
				if ( oCol.sClass )
				{
					nTd.className += ' '+oCol.sClass;
				}
	
				// Visibility - add or remove as required
				if ( oCol.bVisible && ! nTrIn )
				{
					nTr.appendChild( nTd );
				}
				else if ( ! oCol.bVisible && nTrIn )
				{
					nTd.parentNode.removeChild( nTd );
				}
	
				if ( oCol.fnCreatedCell )
				{
					oCol.fnCreatedCell.call( oSettings.oInstance,
						nTd, _fnGetCellData( oSettings, iRow, i ), rowData, iRow, i
					);
				}
			}
	
			_fnCallbackFire( oSettings, 'aoRowCreatedCallback', null, [nTr, rowData, iRow, cells] );
		}
	}
	
	
	/**
	 * Add attributes to a row based on the special `DT_*` parameters in a data
	 * source object.
	 *  @param {object} settings DataTables settings object
	 *  @param {object} DataTables row object for the row to be modified
	 *  @memberof DataTable#oApi
	 */
	function _fnRowAttributes( settings, row )
	{
		var tr = row.nTr;
		var data = row._aData;
	
		if ( tr ) {
			var id = settings.rowIdFn( data );
	
			if ( id ) {
				tr.id = id;
			}
	
			if ( data.DT_RowClass ) {
				// Remove any classes added by DT_RowClass before
				var a = data.DT_RowClass.split(' ');
				row.__rowc = row.__rowc ?
					_unique( row.__rowc.concat( a ) ) :
					a;
	
				$(tr)
					.removeClass( row.__rowc.join(' ') )
					.addClass( data.DT_RowClass );
			}
	
			if ( data.DT_RowAttr ) {
				$(tr).attr( data.DT_RowAttr );
			}
	
			if ( data.DT_RowData ) {
				$(tr).data( data.DT_RowData );
			}
		}
	}
	
	
	/**
	 * Create the HTML header for the table
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnBuildHead( oSettings )
	{
		var i, ien, cell, row, column;
		var thead = oSettings.nTHead;
		var tfoot = oSettings.nTFoot;
		var createHeader = $('th, td', thead).length === 0;
		var classes = oSettings.oClasses;
		var columns = oSettings.aoColumns;
	
		if ( createHeader ) {
			row = $('<tr/>').appendTo( thead );
		}
	
		for ( i=0, ien=columns.length ; i<ien ; i++ ) {
			column = columns[i];
			cell = $( column.nTh ).addClass( column.sClass );
	
			if ( createHeader ) {
				cell.appendTo( row );
			}
	
			// 1.11 move into sorting
			if ( oSettings.oFeatures.bSort ) {
				cell.addClass( column.sSortingClass );
	
				if ( column.bSortable !== false ) {
					cell
						.attr( 'tabindex', oSettings.iTabIndex )
						.attr( 'aria-controls', oSettings.sTableId );
	
					_fnSortAttachListener( oSettings, column.nTh, i );
				}
			}
	
			if ( column.sTitle != cell[0].innerHTML ) {
				cell.html( column.sTitle );
			}
	
			_fnRenderer( oSettings, 'header' )(
				oSettings, cell, column, classes
			);
		}
	
		if ( createHeader ) {
			_fnDetectHeader( oSettings.aoHeader, thead );
		}
	
		/* Deal with the footer - add classes if required */
		$(thead).children('tr').children('th, td').addClass( classes.sHeaderTH );
		$(tfoot).children('tr').children('th, td').addClass( classes.sFooterTH );
	
		// Cache the footer cells. Note that we only take the cells from the first
		// row in the footer. If there is more than one row the user wants to
		// interact with, they need to use the table().foot() method. Note also this
		// allows cells to be used for multiple columns using colspan
		if ( tfoot !== null ) {
			var cells = oSettings.aoFooter[0];
	
			for ( i=0, ien=cells.length ; i<ien ; i++ ) {
				column = columns[i];
	
				if (column) {
					column.nTf = cells[i].cell;
		
					if ( column.sClass ) {
						$(column.nTf).addClass( column.sClass );
					}
				}
				else {
					_fnLog( oSettings, 0, 'Incorrect column count', 18 );
				}
			}
		}
	}
	
	
	/**
	 * Draw the header (or footer) element based on the column visibility states. The
	 * methodology here is to use the layout array from _fnDetectHeader, modified for
	 * the instantaneous column visibility, to construct the new layout. The grid is
	 * traversed over cell at a time in a rows x columns grid fashion, although each
	 * cell insert can cover multiple elements in the grid - which is tracks using the
	 * aApplied array. Cell inserts in the grid will only occur where there isn't
	 * already a cell in that position.
	 *  @param {object} oSettings dataTables settings object
	 *  @param array {objects} aoSource Layout array from _fnDetectHeader
	 *  @param {boolean} [bIncludeHidden=false] If true then include the hidden columns in the calc,
	 *  @memberof DataTable#oApi
	 */
	function _fnDrawHead( oSettings, aoSource, bIncludeHidden )
	{
		var i, iLen, j, jLen, k, kLen, n, nLocalTr;
		var aoLocal = [];
		var aApplied = [];
		var iColumns = oSettings.aoColumns.length;
		var iRowspan, iColspan;
	
		if ( ! aoSource )
		{
			return;
		}
	
		if (  bIncludeHidden === undefined )
		{
			bIncludeHidden = false;
		}
	
		/* Make a copy of the master layout array, but without the visible columns in it */
		for ( i=0, iLen=aoSource.length ; i<iLen ; i++ )
		{
			aoLocal[i] = aoSource[i].slice();
			aoLocal[i].nTr = aoSource[i].nTr;
	
			/* Remove any columns which are currently hidden */
			for ( j=iColumns-1 ; j>=0 ; j-- )
			{
				if ( !oSettings.aoColumns[j].bVisible && !bIncludeHidden )
				{
					aoLocal[i].splice( j, 1 );
				}
			}
	
			/* Prep the applied array - it needs an element for each row */
			aApplied.push( [] );
		}
	
		for ( i=0, iLen=aoLocal.length ; i<iLen ; i++ )
		{
			nLocalTr = aoLocal[i].nTr;
	
			/* All cells are going to be replaced, so empty out the row */
			if ( nLocalTr )
			{
				while( (n = nLocalTr.firstChild) )
				{
					nLocalTr.removeChild( n );
				}
			}
	
			for ( j=0, jLen=aoLocal[i].length ; j<jLen ; j++ )
			{
				iRowspan = 1;
				iColspan = 1;
	
				/* Check to see if there is already a cell (row/colspan) covering our target
				 * insert point. If there is, then there is nothing to do.
				 */
				if ( aApplied[i][j] === undefined )
				{
					nLocalTr.appendChild( aoLocal[i][j].cell );
					aApplied[i][j] = 1;
	
					/* Expand the cell to cover as many rows as needed */
					while ( aoLocal[i+iRowspan] !== undefined &&
					        aoLocal[i][j].cell == aoLocal[i+iRowspan][j].cell )
					{
						aApplied[i+iRowspan][j] = 1;
						iRowspan++;
					}
	
					/* Expand the cell to cover as many columns as needed */
					while ( aoLocal[i][j+iColspan] !== undefined &&
					        aoLocal[i][j].cell == aoLocal[i][j+iColspan].cell )
					{
						/* Must update the applied array over the rows for the columns */
						for ( k=0 ; k<iRowspan ; k++ )
						{
							aApplied[i+k][j+iColspan] = 1;
						}
						iColspan++;
					}
	
					/* Do the actual expansion in the DOM */
					$(aoLocal[i][j].cell)
						.attr('rowspan', iRowspan)
						.attr('colspan', iColspan);
				}
			}
		}
	}
	
	
	/**
	 * Insert the required TR nodes into the table for display
	 *  @param {object} oSettings dataTables settings object
	 *  @param ajaxComplete true after ajax call to complete rendering
	 *  @memberof DataTable#oApi
	 */
	function _fnDraw( oSettings, ajaxComplete )
	{
		// Allow for state saving and a custom start position
		_fnStart( oSettings );
	
		/* Provide a pre-callback function which can be used to cancel the draw is false is returned */
		var aPreDraw = _fnCallbackFire( oSettings, 'aoPreDrawCallback', 'preDraw', [oSettings] );
		if ( $.inArray( false, aPreDraw ) !== -1 )
		{
			_fnProcessingDisplay( oSettings, false );
			return;
		}
	
		var anRows = [];
		var iRowCount = 0;
		var asStripeClasses = oSettings.asStripeClasses;
		var iStripes = asStripeClasses.length;
		var oLang = oSettings.oLanguage;
		var bServerSide = _fnDataSource( oSettings ) == 'ssp';
		var aiDisplay = oSettings.aiDisplay;
		var iDisplayStart = oSettings._iDisplayStart;
		var iDisplayEnd = oSettings.fnDisplayEnd();
	
		oSettings.bDrawing = true;
	
		/* Server-side processing draw intercept */
		if ( oSettings.bDeferLoading )
		{
			oSettings.bDeferLoading = false;
			oSettings.iDraw++;
			_fnProcessingDisplay( oSettings, false );
		}
		else if ( !bServerSide )
		{
			oSettings.iDraw++;
		}
		else if ( !oSettings.bDestroying && !ajaxComplete)
		{
			_fnAjaxUpdate( oSettings );
			return;
		}
	
		if ( aiDisplay.length !== 0 )
		{
			var iStart = bServerSide ? 0 : iDisplayStart;
			var iEnd = bServerSide ? oSettings.aoData.length : iDisplayEnd;
	
			for ( var j=iStart ; j<iEnd ; j++ )
			{
				var iDataIndex = aiDisplay[j];
				var aoData = oSettings.aoData[ iDataIndex ];
				if ( aoData.nTr === null )
				{
					_fnCreateTr( oSettings, iDataIndex );
				}
	
				var nRow = aoData.nTr;
	
				/* Remove the old striping classes and then add the new one */
				if ( iStripes !== 0 )
				{
					var sStripe = asStripeClasses[ iRowCount % iStripes ];
					if ( aoData._sRowStripe != sStripe )
					{
						$(nRow).removeClass( aoData._sRowStripe ).addClass( sStripe );
						aoData._sRowStripe = sStripe;
					}
				}
	
				// Row callback functions - might want to manipulate the row
				// iRowCount and j are not currently documented. Are they at all
				// useful?
				_fnCallbackFire( oSettings, 'aoRowCallback', null,
					[nRow, aoData._aData, iRowCount, j, iDataIndex] );
	
				anRows.push( nRow );
				iRowCount++;
			}
		}
		else
		{
			/* Table is empty - create a row with an empty message in it */
			var sZero = oLang.sZeroRecords;
			if ( oSettings.iDraw == 1 &&  _fnDataSource( oSettings ) == 'ajax' )
			{
				sZero = oLang.sLoadingRecords;
			}
			else if ( oLang.sEmptyTable && oSettings.fnRecordsTotal() === 0 )
			{
				sZero = oLang.sEmptyTable;
			}
	
			anRows[ 0 ] = $( '<tr/>', { 'class': iStripes ? asStripeClasses[0] : '' } )
				.append( $('<td />', {
					'valign':  'top',
					'colSpan': _fnVisbleColumns( oSettings ),
					'class':   oSettings.oClasses.sRowEmpty
				} ).html( sZero ) )[0];
		}
	
		/* Header and footer callbacks */
		_fnCallbackFire( oSettings, 'aoHeaderCallback', 'header', [ $(oSettings.nTHead).children('tr')[0],
			_fnGetDataMaster( oSettings ), iDisplayStart, iDisplayEnd, aiDisplay ] );
	
		_fnCallbackFire( oSettings, 'aoFooterCallback', 'footer', [ $(oSettings.nTFoot).children('tr')[0],
			_fnGetDataMaster( oSettings ), iDisplayStart, iDisplayEnd, aiDisplay ] );
	
		var body = $(oSettings.nTBody);
	
		body.children().detach();
		body.append( $(anRows) );
	
		/* Call all required callback functions for the end of a draw */
		_fnCallbackFire( oSettings, 'aoDrawCallback', 'draw', [oSettings] );
	
		/* Draw is complete, sorting and filtering must be as well */
		oSettings.bSorted = false;
		oSettings.bFiltered = false;
		oSettings.bDrawing = false;
	}
	
	
	/**
	 * Redraw the table - taking account of the various features which are enabled
	 *  @param {object} oSettings dataTables settings object
	 *  @param {boolean} [holdPosition] Keep the current paging position. By default
	 *    the paging is reset to the first page
	 *  @memberof DataTable#oApi
	 */
	function _fnReDraw( settings, holdPosition )
	{
		var
			features = settings.oFeatures,
			sort     = features.bSort,
			filter   = features.bFilter;
	
		if ( sort ) {
			_fnSort( settings );
		}
	
		if ( filter ) {
			_fnFilterComplete( settings, settings.oPreviousSearch );
		}
		else {
			// No filtering, so we want to just use the display master
			settings.aiDisplay = settings.aiDisplayMaster.slice();
		}
	
		if ( holdPosition !== true ) {
			settings._iDisplayStart = 0;
		}
	
		// Let any modules know about the draw hold position state (used by
		// scrolling internally)
		settings._drawHold = holdPosition;
	
		_fnDraw( settings );
	
		settings._drawHold = false;
	}
	
	
	/**
	 * Add the options to the page HTML for the table
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnAddOptionsHtml ( oSettings )
	{
		var classes = oSettings.oClasses;
		var table = $(oSettings.nTable);
		var holding = $('<div/>').insertBefore( table ); // Holding element for speed
		var features = oSettings.oFeatures;
	
		// All DataTables are wrapped in a div
		var insert = $('<div/>', {
			id:      oSettings.sTableId+'_wrapper',
			'class': classes.sWrapper + (oSettings.nTFoot ? '' : ' '+classes.sNoFooter)
		} );
	
		oSettings.nHolding = holding[0];
		oSettings.nTableWrapper = insert[0];
		oSettings.nTableReinsertBefore = oSettings.nTable.nextSibling;
	
		/* Loop over the user set positioning and place the elements as needed */
		var aDom = oSettings.sDom.split('');
		var featureNode, cOption, nNewNode, cNext, sAttr, j;
		for ( var i=0 ; i<aDom.length ; i++ )
		{
			featureNode = null;
			cOption = aDom[i];
	
			if ( cOption == '<' )
			{
				/* New container div */
				nNewNode = $('<div/>')[0];
	
				/* Check to see if we should append an id and/or a class name to the container */
				cNext = aDom[i+1];
				if ( cNext == "'" || cNext == '"' )
				{
					sAttr = "";
					j = 2;
					while ( aDom[i+j] != cNext )
					{
						sAttr += aDom[i+j];
						j++;
					}
	
					/* Replace jQuery UI constants @todo depreciated */
					if ( sAttr == "H" )
					{
						sAttr = classes.sJUIHeader;
					}
					else if ( sAttr == "F" )
					{
						sAttr = classes.sJUIFooter;
					}
	
					/* The attribute can be in the format of "#id.class", "#id" or "class" This logic
					 * breaks the string into parts and applies them as needed
					 */
					if ( sAttr.indexOf('.') != -1 )
					{
						var aSplit = sAttr.split('.');
						nNewNode.id = aSplit[0].substr(1, aSplit[0].length-1);
						nNewNode.className = aSplit[1];
					}
					else if ( sAttr.charAt(0) == "#" )
					{
						nNewNode.id = sAttr.substr(1, sAttr.length-1);
					}
					else
					{
						nNewNode.className = sAttr;
					}
	
					i += j; /* Move along the position array */
				}
	
				insert.append( nNewNode );
				insert = $(nNewNode);
			}
			else if ( cOption == '>' )
			{
				/* End container div */
				insert = insert.parent();
			}
			// @todo Move options into their own plugins?
			else if ( cOption == 'l' && features.bPaginate && features.bLengthChange )
			{
				/* Length */
				featureNode = _fnFeatureHtmlLength( oSettings );
			}
			else if ( cOption == 'f' && features.bFilter )
			{
				/* Filter */
				featureNode = _fnFeatureHtmlFilter( oSettings );
			}
			else if ( cOption == 'r' && features.bProcessing )
			{
				/* pRocessing */
				featureNode = _fnFeatureHtmlProcessing( oSettings );
			}
			else if ( cOption == 't' )
			{
				/* Table */
				featureNode = _fnFeatureHtmlTable( oSettings );
			}
			else if ( cOption ==  'i' && features.bInfo )
			{
				/* Info */
				featureNode = _fnFeatureHtmlInfo( oSettings );
			}
			else if ( cOption == 'p' && features.bPaginate )
			{
				/* Pagination */
				featureNode = _fnFeatureHtmlPaginate( oSettings );
			}
			else if ( DataTable.ext.feature.length !== 0 )
			{
				/* Plug-in features */
				var aoFeatures = DataTable.ext.feature;
				for ( var k=0, kLen=aoFeatures.length ; k<kLen ; k++ )
				{
					if ( cOption == aoFeatures[k].cFeature )
					{
						featureNode = aoFeatures[k].fnInit( oSettings );
						break;
					}
				}
			}
	
			/* Add to the 2D features array */
			if ( featureNode )
			{
				var aanFeatures = oSettings.aanFeatures;
	
				if ( ! aanFeatures[cOption] )
				{
					aanFeatures[cOption] = [];
				}
	
				aanFeatures[cOption].push( featureNode );
				insert.append( featureNode );
			}
		}
	
		/* Built our DOM structure - replace the holding div with what we want */
		holding.replaceWith( insert );
		oSettings.nHolding = null;
	}
	
	
	/**
	 * Use the DOM source to create up an array of header cells. The idea here is to
	 * create a layout grid (array) of rows x columns, which contains a reference
	 * to the cell that that point in the grid (regardless of col/rowspan), such that
	 * any column / row could be removed and the new grid constructed
	 *  @param array {object} aLayout Array to store the calculated layout in
	 *  @param {node} nThead The header/footer element for the table
	 *  @memberof DataTable#oApi
	 */
	function _fnDetectHeader ( aLayout, nThead )
	{
		var nTrs = $(nThead).children('tr');
		var nTr, nCell;
		var i, k, l, iLen, jLen, iColShifted, iColumn, iColspan, iRowspan;
		var bUnique;
		var fnShiftCol = function ( a, i, j ) {
			var k = a[i];
	                while ( k[j] ) {
				j++;
			}
			return j;
		};
	
		aLayout.splice( 0, aLayout.length );
	
		/* We know how many rows there are in the layout - so prep it */
		for ( i=0, iLen=nTrs.length ; i<iLen ; i++ )
		{
			aLayout.push( [] );
		}
	
		/* Calculate a layout array */
		for ( i=0, iLen=nTrs.length ; i<iLen ; i++ )
		{
			nTr = nTrs[i];
			iColumn = 0;
	
			/* For every cell in the row... */
			nCell = nTr.firstChild;
			while ( nCell ) {
				if ( nCell.nodeName.toUpperCase() == "TD" ||
				     nCell.nodeName.toUpperCase() == "TH" )
				{
					/* Get the col and rowspan attributes from the DOM and sanitise them */
					iColspan = nCell.getAttribute('colspan') * 1;
					iRowspan = nCell.getAttribute('rowspan') * 1;
					iColspan = (!iColspan || iColspan===0 || iColspan===1) ? 1 : iColspan;
					iRowspan = (!iRowspan || iRowspan===0 || iRowspan===1) ? 1 : iRowspan;
	
					/* There might be colspan cells already in this row, so shift our target
					 * accordingly
					 */
					iColShifted = fnShiftCol( aLayout, i, iColumn );
	
					/* Cache calculation for unique columns */
					bUnique = iColspan === 1 ? true : false;
	
					/* If there is col / rowspan, copy the information into the layout grid */
					for ( l=0 ; l<iColspan ; l++ )
					{
						for ( k=0 ; k<iRowspan ; k++ )
						{
							aLayout[i+k][iColShifted+l] = {
								"cell": nCell,
								"unique": bUnique
							};
							aLayout[i+k].nTr = nTr;
						}
					}
				}
				nCell = nCell.nextSibling;
			}
		}
	}
	
	
	/**
	 * Get an array of unique th elements, one for each column
	 *  @param {object} oSettings dataTables settings object
	 *  @param {node} nHeader automatically detect the layout from this node - optional
	 *  @param {array} aLayout thead/tfoot layout from _fnDetectHeader - optional
	 *  @returns array {node} aReturn list of unique th's
	 *  @memberof DataTable#oApi
	 */
	function _fnGetUniqueThs ( oSettings, nHeader, aLayout )
	{
		var aReturn = [];
		if ( !aLayout )
		{
			aLayout = oSettings.aoHeader;
			if ( nHeader )
			{
				aLayout = [];
				_fnDetectHeader( aLayout, nHeader );
			}
		}
	
		for ( var i=0, iLen=aLayout.length ; i<iLen ; i++ )
		{
			for ( var j=0, jLen=aLayout[i].length ; j<jLen ; j++ )
			{
				if ( aLayout[i][j].unique &&
					 (!aReturn[j] || !oSettings.bSortCellsTop) )
				{
					aReturn[j] = aLayout[i][j].cell;
				}
			}
		}
	
		return aReturn;
	}
	
	/**
	 * Set the start position for draw
	 *  @param {object} oSettings dataTables settings object
	 */
	function _fnStart( oSettings )
	{
		var bServerSide = _fnDataSource( oSettings ) == 'ssp';
		var iInitDisplayStart = oSettings.iInitDisplayStart;
	
		// Check and see if we have an initial draw position from state saving
		if ( iInitDisplayStart !== undefined && iInitDisplayStart !== -1 )
		{
			oSettings._iDisplayStart = bServerSide ?
				iInitDisplayStart :
				iInitDisplayStart >= oSettings.fnRecordsDisplay() ?
					0 :
					iInitDisplayStart;
	
			oSettings.iInitDisplayStart = -1;
		}
	}
	
	/**
	 * Create an Ajax call based on the table's settings, taking into account that
	 * parameters can have multiple forms, and backwards compatibility.
	 *
	 * @param {object} oSettings dataTables settings object
	 * @param {array} data Data to send to the server, required by
	 *     DataTables - may be augmented by developer callbacks
	 * @param {function} fn Callback function to run when data is obtained
	 */
	function _fnBuildAjax( oSettings, data, fn )
	{
		// Compatibility with 1.9-, allow fnServerData and event to manipulate
		_fnCallbackFire( oSettings, 'aoServerParams', 'serverParams', [data] );
	
		// Convert to object based for 1.10+ if using the old array scheme which can
		// come from server-side processing or serverParams
		if ( data && Array.isArray(data) ) {
			var tmp = {};
			var rbracket = /(.*?)\[\]$/;
	
			$.each( data, function (key, val) {
				var match = val.name.match(rbracket);
	
				if ( match ) {
					// Support for arrays
					var name = match[0];
	
					if ( ! tmp[ name ] ) {
						tmp[ name ] = [];
					}
					tmp[ name ].push( val.value );
				}
				else {
					tmp[val.name] = val.value;
				}
			} );
			data = tmp;
		}
	
		var ajaxData;
		var ajax = oSettings.ajax;
		var instance = oSettings.oInstance;
		var callback = function ( json ) {
			var status = oSettings.jqXHR
				? oSettings.jqXHR.status
				: null;
	
			if ( json === null || (typeof status === 'number' && status == 204 ) ) {
				json = {};
				_fnAjaxDataSrc( oSettings, json, [] );
			}
	
			var error = json.error || json.sError;
			if ( error ) {
				_fnLog( oSettings, 0, error );
			}
	
			oSettings.json = json;
	
			_fnCallbackFire( oSettings, null, 'xhr', [oSettings, json, oSettings.jqXHR] );
			fn( json );
		};
	
		if ( $.isPlainObject( ajax ) && ajax.data )
		{
			ajaxData = ajax.data;
	
			var newData = typeof ajaxData === 'function' ?
				ajaxData( data, oSettings ) :  // fn can manipulate data or return
				ajaxData;                      // an object object or array to merge
	
			// If the function returned something, use that alone
			data = typeof ajaxData === 'function' && newData ?
				newData :
				$.extend( true, data, newData );
	
			// Remove the data property as we've resolved it already and don't want
			// jQuery to do it again (it is restored at the end of the function)
			delete ajax.data;
		}
	
		var baseAjax = {
			"data": data,
			"success": callback,
			"dataType": "json",
			"cache": false,
			"type": oSettings.sServerMethod,
			"error": function (xhr, error, thrown) {
				var ret = _fnCallbackFire( oSettings, null, 'xhr', [oSettings, null, oSettings.jqXHR] );
	
				if ( $.inArray( true, ret ) === -1 ) {
					if ( error == "parsererror" ) {
						_fnLog( oSettings, 0, 'Invalid JSON response', 1 );
					}
					else if ( xhr.readyState === 4 ) {
						_fnLog( oSettings, 0, 'Ajax error', 7 );
					}
				}
	
				_fnProcessingDisplay( oSettings, false );
			}
		};
	
		// Store the data submitted for the API
		oSettings.oAjaxData = data;
	
		// Allow plug-ins and external processes to modify the data
		_fnCallbackFire( oSettings, null, 'preXhr', [oSettings, data] );
	
		if ( oSettings.fnServerData )
		{
			// DataTables 1.9- compatibility
			oSettings.fnServerData.call( instance,
				oSettings.sAjaxSource,
				$.map( data, function (val, key) { // Need to convert back to 1.9 trad format
					return { name: key, value: val };
				} ),
				callback,
				oSettings
			);
		}
		else if ( oSettings.sAjaxSource || typeof ajax === 'string' )
		{
			// DataTables 1.9- compatibility
			oSettings.jqXHR = $.ajax( $.extend( baseAjax, {
				url: ajax || oSettings.sAjaxSource
			} ) );
		}
		else if ( typeof ajax === 'function' )
		{
			// Is a function - let the caller define what needs to be done
			oSettings.jqXHR = ajax.call( instance, data, callback, oSettings );
		}
		else
		{
			// Object to extend the base settings
			oSettings.jqXHR = $.ajax( $.extend( baseAjax, ajax ) );
	
			// Restore for next time around
			ajax.data = ajaxData;
		}
	}
	
	
	/**
	 * Update the table using an Ajax call
	 *  @param {object} settings dataTables settings object
	 *  @returns {boolean} Block the table drawing or not
	 *  @memberof DataTable#oApi
	 */
	function _fnAjaxUpdate( settings )
	{
		settings.iDraw++;
		_fnProcessingDisplay( settings, true );
	
		_fnBuildAjax(
			settings,
			_fnAjaxParameters( settings ),
			function(json) {
				_fnAjaxUpdateDraw( settings, json );
			}
		);
	}
	
	
	/**
	 * Build up the parameters in an object needed for a server-side processing
	 * request. Note that this is basically done twice, is different ways - a modern
	 * method which is used by default in DataTables 1.10 which uses objects and
	 * arrays, or the 1.9- method with is name / value pairs. 1.9 method is used if
	 * the sAjaxSource option is used in the initialisation, or the legacyAjax
	 * option is set.
	 *  @param {object} oSettings dataTables settings object
	 *  @returns {bool} block the table drawing or not
	 *  @memberof DataTable#oApi
	 */
	function _fnAjaxParameters( settings )
	{
		var
			columns = settings.aoColumns,
			columnCount = columns.length,
			features = settings.oFeatures,
			preSearch = settings.oPreviousSearch,
			preColSearch = settings.aoPreSearchCols,
			i, data = [], dataProp, column, columnSearch,
			sort = _fnSortFlatten( settings ),
			displayStart = settings._iDisplayStart,
			displayLength = features.bPaginate !== false ?
				settings._iDisplayLength :
				-1;
	
		var param = function ( name, value ) {
			data.push( { 'name': name, 'value': value } );
		};
	
		// DataTables 1.9- compatible method
		param( 'sEcho',          settings.iDraw );
		param( 'iColumns',       columnCount );
		param( 'sColumns',       _pluck( columns, 'sName' ).join(',') );
		param( 'iDisplayStart',  displayStart );
		param( 'iDisplayLength', displayLength );
	
		// DataTables 1.10+ method
		var d = {
			draw:    settings.iDraw,
			columns: [],
			order:   [],
			start:   displayStart,
			length:  displayLength,
			search:  {
				value: preSearch.sSearch,
				regex: preSearch.bRegex
			}
		};
	
		for ( i=0 ; i<columnCount ; i++ ) {
			column = columns[i];
			columnSearch = preColSearch[i];
			dataProp = typeof column.mData=="function" ? 'function' : column.mData ;
	
			d.columns.push( {
				data:       dataProp,
				name:       column.sName,
				searchable: column.bSearchable,
				orderable:  column.bSortable,
				search:     {
					value: columnSearch.sSearch,
					regex: columnSearch.bRegex
				}
			} );
	
			param( "mDataProp_"+i, dataProp );
	
			if ( features.bFilter ) {
				param( 'sSearch_'+i,     columnSearch.sSearch );
				param( 'bRegex_'+i,      columnSearch.bRegex );
				param( 'bSearchable_'+i, column.bSearchable );
			}
	
			if ( features.bSort ) {
				param( 'bSortable_'+i, column.bSortable );
			}
		}
	
		if ( features.bFilter ) {
			param( 'sSearch', preSearch.sSearch );
			param( 'bRegex', preSearch.bRegex );
		}
	
		if ( features.bSort ) {
			$.each( sort, function ( i, val ) {
				d.order.push( { column: val.col, dir: val.dir } );
	
				param( 'iSortCol_'+i, val.col );
				param( 'sSortDir_'+i, val.dir );
			} );
	
			param( 'iSortingCols', sort.length );
		}
	
		// If the legacy.ajax parameter is null, then we automatically decide which
		// form to use, based on sAjaxSource
		var legacy = DataTable.ext.legacy.ajax;
		if ( legacy === null ) {
			return settings.sAjaxSource ? data : d;
		}
	
		// Otherwise, if legacy has been specified then we use that to decide on the
		// form
		return legacy ? data : d;
	}
	
	
	/**
	 * Data the data from the server (nuking the old) and redraw the table
	 *  @param {object} oSettings dataTables settings object
	 *  @param {object} json json data return from the server.
	 *  @param {string} json.sEcho Tracking flag for DataTables to match requests
	 *  @param {int} json.iTotalRecords Number of records in the data set, not accounting for filtering
	 *  @param {int} json.iTotalDisplayRecords Number of records in the data set, accounting for filtering
	 *  @param {array} json.aaData The data to display on this page
	 *  @param {string} [json.sColumns] Column ordering (sName, comma separated)
	 *  @memberof DataTable#oApi
	 */
	function _fnAjaxUpdateDraw ( settings, json )
	{
		// v1.10 uses camelCase variables, while 1.9 uses Hungarian notation.
		// Support both
		var compat = function ( old, modern ) {
			return json[old] !== undefined ? json[old] : json[modern];
		};
	
		var data = _fnAjaxDataSrc( settings, json );
		var draw            = compat( 'sEcho',                'draw' );
		var recordsTotal    = compat( 'iTotalRecords',        'recordsTotal' );
		var recordsFiltered = compat( 'iTotalDisplayRecords', 'recordsFiltered' );
	
		if ( draw !== undefined ) {
			// Protect against out of sequence returns
			if ( draw*1 < settings.iDraw ) {
				return;
			}
			settings.iDraw = draw * 1;
		}
	
		// No data in returned object, so rather than an array, we show an empty table
		if ( ! data ) {
			data = [];
		}
	
		_fnClearTable( settings );
		settings._iRecordsTotal   = parseInt(recordsTotal, 10);
		settings._iRecordsDisplay = parseInt(recordsFiltered, 10);
	
		for ( var i=0, ien=data.length ; i<ien ; i++ ) {
			_fnAddData( settings, data[i] );
		}
		settings.aiDisplay = settings.aiDisplayMaster.slice();
	
		_fnDraw( settings, true );
	
		if ( ! settings._bInitComplete ) {
			_fnInitComplete( settings, json );
		}
	
		_fnProcessingDisplay( settings, false );
	}
	
	
	/**
	 * Get the data from the JSON data source to use for drawing a table. Using
	 * `_fnGetObjectDataFn` allows the data to be sourced from a property of the
	 * source object, or from a processing function.
	 *  @param {object} oSettings dataTables settings object
	 *  @param  {object} json Data source object / array from the server
	 *  @return {array} Array of data to use
	 */
	 function _fnAjaxDataSrc ( oSettings, json, write )
	 {
		var dataSrc = $.isPlainObject( oSettings.ajax ) && oSettings.ajax.dataSrc !== undefined ?
			oSettings.ajax.dataSrc :
			oSettings.sAjaxDataProp; // Compatibility with 1.9-.
	
		if ( ! write ) {
			if ( dataSrc === 'data' ) {
				// If the default, then we still want to support the old style, and safely ignore
				// it if possible
				return json.aaData || json[dataSrc];
			}
	
			return dataSrc !== "" ?
				_fnGetObjectDataFn( dataSrc )( json ) :
				json;
		}
	
		// set
		_fnSetObjectDataFn( dataSrc )( json, write );
	}
	
	/**
	 * Generate the node required for filtering text
	 *  @returns {node} Filter control element
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnFeatureHtmlFilter ( settings )
	{
		var classes = settings.oClasses;
		var tableId = settings.sTableId;
		var language = settings.oLanguage;
		var previousSearch = settings.oPreviousSearch;
		var features = settings.aanFeatures;
		var input = '<input type="search" class="'+classes.sFilterInput+'"/>';
	
		var str = language.sSearch;
		str = str.match(/_INPUT_/) ?
			str.replace('_INPUT_', input) :
			str+input;
	
		var filter = $('<div/>', {
				'id': ! features.f ? tableId+'_filter' : null,
				'class': classes.sFilter
			} )
			.append( $('<label/>' ).append( str ) );
	
		var searchFn = function(event) {
			/* Update all other filter input elements for the new display */
			var n = features.f;
			var val = !this.value ? "" : this.value; // mental IE8 fix :-(
			if(previousSearch.return && event.key !== "Enter") {
				return;
			}
			/* Now do the filter */
			if ( val != previousSearch.sSearch ) {
				_fnFilterComplete( settings, {
					"sSearch": val,
					"bRegex": previousSearch.bRegex,
					"bSmart": previousSearch.bSmart ,
					"bCaseInsensitive": previousSearch.bCaseInsensitive,
					"return": previousSearch.return
				} );
	
				// Need to redraw, without resorting
				settings._iDisplayStart = 0;
				_fnDraw( settings );
			}
		};
	
		var searchDelay = settings.searchDelay !== null ?
			settings.searchDelay :
			_fnDataSource( settings ) === 'ssp' ?
				400 :
				0;
	
		var jqFilter = $('input', filter)
			.val( previousSearch.sSearch )
			.attr( 'placeholder', language.sSearchPlaceholder )
			.on(
				'keyup.DT search.DT input.DT paste.DT cut.DT',
				searchDelay ?
					_fnThrottle( searchFn, searchDelay ) :
					searchFn
			)
			.on( 'mouseup', function(e) {
				// Edge fix! Edge 17 does not trigger anything other than mouse events when clicking
				// on the clear icon (Edge bug 17584515). This is safe in other browsers as `searchFn`
				// checks the value to see if it has changed. In other browsers it won't have.
				setTimeout( function () {
					searchFn.call(jqFilter[0], e);
				}, 10);
			} )
			.on( 'keypress.DT', function(e) {
				/* Prevent form submission */
				if ( e.keyCode == 13 ) {
					return false;
				}
			} )
			.attr('aria-controls', tableId);
	
		// Update the input elements whenever the table is filtered
		$(settings.nTable).on( 'search.dt.DT', function ( ev, s ) {
			if ( settings === s ) {
				// IE9 throws an 'unknown error' if document.activeElement is used
				// inside an iframe or frame...
				try {
					if ( jqFilter[0] !== document.activeElement ) {
						jqFilter.val( previousSearch.sSearch );
					}
				}
				catch ( e ) {}
			}
		} );
	
		return filter[0];
	}
	
	
	/**
	 * Filter the table using both the global filter and column based filtering
	 *  @param {object} oSettings dataTables settings object
	 *  @param {object} oSearch search information
	 *  @param {int} [iForce] force a research of the master array (1) or not (undefined or 0)
	 *  @memberof DataTable#oApi
	 */
	function _fnFilterComplete ( oSettings, oInput, iForce )
	{
		var oPrevSearch = oSettings.oPreviousSearch;
		var aoPrevSearch = oSettings.aoPreSearchCols;
		var fnSaveFilter = function ( oFilter ) {
			/* Save the filtering values */
			oPrevSearch.sSearch = oFilter.sSearch;
			oPrevSearch.bRegex = oFilter.bRegex;
			oPrevSearch.bSmart = oFilter.bSmart;
			oPrevSearch.bCaseInsensitive = oFilter.bCaseInsensitive;
			oPrevSearch.return = oFilter.return;
		};
		var fnRegex = function ( o ) {
			// Backwards compatibility with the bEscapeRegex option
			return o.bEscapeRegex !== undefined ? !o.bEscapeRegex : o.bRegex;
		};
	
		// Resolve any column types that are unknown due to addition or invalidation
		// @todo As per sort - can this be moved into an event handler?
		_fnColumnTypes( oSettings );
	
		/* In server-side processing all filtering is done by the server, so no point hanging around here */
		if ( _fnDataSource( oSettings ) != 'ssp' )
		{
			/* Global filter */
			_fnFilter( oSettings, oInput.sSearch, iForce, fnRegex(oInput), oInput.bSmart, oInput.bCaseInsensitive, oInput.return );
			fnSaveFilter( oInput );
	
			/* Now do the individual column filter */
			for ( var i=0 ; i<aoPrevSearch.length ; i++ )
			{
				_fnFilterColumn( oSettings, aoPrevSearch[i].sSearch, i, fnRegex(aoPrevSearch[i]),
					aoPrevSearch[i].bSmart, aoPrevSearch[i].bCaseInsensitive );
			}
	
			/* Custom filtering */
			_fnFilterCustom( oSettings );
		}
		else
		{
			fnSaveFilter( oInput );
		}
	
		/* Tell the draw function we have been filtering */
		oSettings.bFiltered = true;
		_fnCallbackFire( oSettings, null, 'search', [oSettings] );
	}
	
	
	/**
	 * Apply custom filtering functions
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnFilterCustom( settings )
	{
		var filters = DataTable.ext.search;
		var displayRows = settings.aiDisplay;
		var row, rowIdx;
	
		for ( var i=0, ien=filters.length ; i<ien ; i++ ) {
			var rows = [];
	
			// Loop over each row and see if it should be included
			for ( var j=0, jen=displayRows.length ; j<jen ; j++ ) {
				rowIdx = displayRows[ j ];
				row = settings.aoData[ rowIdx ];
	
				if ( filters[i]( settings, row._aFilterData, rowIdx, row._aData, j ) ) {
					rows.push( rowIdx );
				}
			}
	
			// So the array reference doesn't break set the results into the
			// existing array
			displayRows.length = 0;
			$.merge( displayRows, rows );
		}
	}
	
	
	/**
	 * Filter the table on a per-column basis
	 *  @param {object} oSettings dataTables settings object
	 *  @param {string} sInput string to filter on
	 *  @param {int} iColumn column to filter
	 *  @param {bool} bRegex treat search string as a regular expression or not
	 *  @param {bool} bSmart use smart filtering or not
	 *  @param {bool} bCaseInsensitive Do case insensitive matching or not
	 *  @memberof DataTable#oApi
	 */
	function _fnFilterColumn ( settings, searchStr, colIdx, regex, smart, caseInsensitive )
	{
		if ( searchStr === '' ) {
			return;
		}
	
		var data;
		var out = [];
		var display = settings.aiDisplay;
		var rpSearch = _fnFilterCreateSearch( searchStr, regex, smart, caseInsensitive );
	
		for ( var i=0 ; i<display.length ; i++ ) {
			data = settings.aoData[ display[i] ]._aFilterData[ colIdx ];
	
			if ( rpSearch.test( data ) ) {
				out.push( display[i] );
			}
		}
	
		settings.aiDisplay = out;
	}
	
	
	/**
	 * Filter the data table based on user input and draw the table
	 *  @param {object} settings dataTables settings object
	 *  @param {string} input string to filter on
	 *  @param {int} force optional - force a research of the master array (1) or not (undefined or 0)
	 *  @param {bool} regex treat as a regular expression or not
	 *  @param {bool} smart perform smart filtering or not
	 *  @param {bool} caseInsensitive Do case insensitive matching or not
	 *  @memberof DataTable#oApi
	 */
	function _fnFilter( settings, input, force, regex, smart, caseInsensitive )
	{
		var rpSearch = _fnFilterCreateSearch( input, regex, smart, caseInsensitive );
		var prevSearch = settings.oPreviousSearch.sSearch;
		var displayMaster = settings.aiDisplayMaster;
		var display, invalidated, i;
		var filtered = [];
	
		// Need to take account of custom filtering functions - always filter
		if ( DataTable.ext.search.length !== 0 ) {
			force = true;
		}
	
		// Check if any of the rows were invalidated
		invalidated = _fnFilterData( settings );
	
		// If the input is blank - we just want the full data set
		if ( input.length <= 0 ) {
			settings.aiDisplay = displayMaster.slice();
		}
		else {
			// New search - start from the master array
			if ( invalidated ||
				 force ||
				 regex ||
				 prevSearch.length > input.length ||
				 input.indexOf(prevSearch) !== 0 ||
				 settings.bSorted // On resort, the display master needs to be
				                  // re-filtered since indexes will have changed
			) {
				settings.aiDisplay = displayMaster.slice();
			}
	
			// Search the display array
			display = settings.aiDisplay;
	
			for ( i=0 ; i<display.length ; i++ ) {
				if ( rpSearch.test( settings.aoData[ display[i] ]._sFilterRow ) ) {
					filtered.push( display[i] );
				}
			}
	
			settings.aiDisplay = filtered;
		}
	}
	
	
	/**
	 * Build a regular expression object suitable for searching a table
	 *  @param {string} sSearch string to search for
	 *  @param {bool} bRegex treat as a regular expression or not
	 *  @param {bool} bSmart perform smart filtering or not
	 *  @param {bool} bCaseInsensitive Do case insensitive matching or not
	 *  @returns {RegExp} constructed object
	 *  @memberof DataTable#oApi
	 */
	function _fnFilterCreateSearch( search, regex, smart, caseInsensitive )
	{
		search = regex ?
			search :
			_fnEscapeRegex( search );
		
		if ( smart ) {
			/* For smart filtering we want to allow the search to work regardless of
			 * word order. We also want double quoted text to be preserved, so word
			 * order is important - a la google. So this is what we want to
			 * generate:
			 * 
			 * ^(?=.*?\bone\b)(?=.*?\btwo three\b)(?=.*?\bfour\b).*$
			 */
			var a = $.map( search.match( /"[^"]+"|[^ ]+/g ) || [''], function ( word ) {
				if ( word.charAt(0) === '"' ) {
					var m = word.match( /^"(.*)"$/ );
					word = m ? m[1] : word;
				}
	
				return word.replace('"', '');
			} );
	
			search = '^(?=.*?'+a.join( ')(?=.*?' )+').*$';
		}
	
		return new RegExp( search, caseInsensitive ? 'i' : '' );
	}
	
	
	/**
	 * Escape a string such that it can be used in a regular expression
	 *  @param {string} sVal string to escape
	 *  @returns {string} escaped string
	 *  @memberof DataTable#oApi
	 */
	var _fnEscapeRegex = DataTable.util.escapeRegex;
	
	var __filter_div = $('<div>')[0];
	var __filter_div_textContent = __filter_div.textContent !== undefined;
	
	// Update the filtering data for each row if needed (by invalidation or first run)
	function _fnFilterData ( settings )
	{
		var columns = settings.aoColumns;
		var column;
		var i, j, ien, jen, filterData, cellData, row;
		var wasInvalidated = false;
	
		for ( i=0, ien=settings.aoData.length ; i<ien ; i++ ) {
			row = settings.aoData[i];
	
			if ( ! row._aFilterData ) {
				filterData = [];
	
				for ( j=0, jen=columns.length ; j<jen ; j++ ) {
					column = columns[j];
	
					if ( column.bSearchable ) {
						cellData = _fnGetCellData( settings, i, j, 'filter' );
	
						// Search in DataTables 1.10 is string based. In 1.11 this
						// should be altered to also allow strict type checking.
						if ( cellData === null ) {
							cellData = '';
						}
	
						if ( typeof cellData !== 'string' && cellData.toString ) {
							cellData = cellData.toString();
						}
					}
					else {
						cellData = '';
					}
	
					// If it looks like there is an HTML entity in the string,
					// attempt to decode it so sorting works as expected. Note that
					// we could use a single line of jQuery to do this, but the DOM
					// method used here is much faster http://jsperf.com/html-decode
					if ( cellData.indexOf && cellData.indexOf('&') !== -1 ) {
						__filter_div.innerHTML = cellData;
						cellData = __filter_div_textContent ?
							__filter_div.textContent :
							__filter_div.innerText;
					}
	
					if ( cellData.replace ) {
						cellData = cellData.replace(/[\r\n\u2028]/g, '');
					}
	
					filterData.push( cellData );
				}
	
				row._aFilterData = filterData;
				row._sFilterRow = filterData.join('  ');
				wasInvalidated = true;
			}
		}
	
		return wasInvalidated;
	}
	
	
	/**
	 * Convert from the internal Hungarian notation to camelCase for external
	 * interaction
	 *  @param {object} obj Object to convert
	 *  @returns {object} Inverted object
	 *  @memberof DataTable#oApi
	 */
	function _fnSearchToCamel ( obj )
	{
		return {
			search:          obj.sSearch,
			smart:           obj.bSmart,
			regex:           obj.bRegex,
			caseInsensitive: obj.bCaseInsensitive
		};
	}
	
	
	
	/**
	 * Convert from camelCase notation to the internal Hungarian. We could use the
	 * Hungarian convert function here, but this is cleaner
	 *  @param {object} obj Object to convert
	 *  @returns {object} Inverted object
	 *  @memberof DataTable#oApi
	 */
	function _fnSearchToHung ( obj )
	{
		return {
			sSearch:          obj.search,
			bSmart:           obj.smart,
			bRegex:           obj.regex,
			bCaseInsensitive: obj.caseInsensitive
		};
	}
	
	/**
	 * Generate the node required for the info display
	 *  @param {object} oSettings dataTables settings object
	 *  @returns {node} Information element
	 *  @memberof DataTable#oApi
	 */
	function _fnFeatureHtmlInfo ( settings )
	{
		var
			tid = settings.sTableId,
			nodes = settings.aanFeatures.i,
			n = $('<div/>', {
				'class': settings.oClasses.sInfo,
				'id': ! nodes ? tid+'_info' : null
			} );
	
		if ( ! nodes ) {
			// Update display on each draw
			settings.aoDrawCallback.push( {
				"fn": _fnUpdateInfo,
				"sName": "information"
			} );
	
			n
				.attr( 'role', 'status' )
				.attr( 'aria-live', 'polite' );
	
			// Table is described by our info div
			$(settings.nTable).attr( 'aria-describedby', tid+'_info' );
		}
	
		return n[0];
	}
	
	
	/**
	 * Update the information elements in the display
	 *  @param {object} settings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnUpdateInfo ( settings )
	{
		/* Show information about the table */
		var nodes = settings.aanFeatures.i;
		if ( nodes.length === 0 ) {
			return;
		}
	
		var
			lang  = settings.oLanguage,
			start = settings._iDisplayStart+1,
			end   = settings.fnDisplayEnd(),
			max   = settings.fnRecordsTotal(),
			total = settings.fnRecordsDisplay(),
			out   = total ?
				lang.sInfo :
				lang.sInfoEmpty;
	
		if ( total !== max ) {
			/* Record set after filtering */
			out += ' ' + lang.sInfoFiltered;
		}
	
		// Convert the macros
		out += lang.sInfoPostFix;
		out = _fnInfoMacros( settings, out );
	
		var callback = lang.fnInfoCallback;
		if ( callback !== null ) {
			out = callback.call( settings.oInstance,
				settings, start, end, max, total, out
			);
		}
	
		$(nodes).html( out );
	}
	
	
	function _fnInfoMacros ( settings, str )
	{
		// When infinite scrolling, we are always starting at 1. _iDisplayStart is used only
		// internally
		var
			formatter  = settings.fnFormatNumber,
			start      = settings._iDisplayStart+1,
			len        = settings._iDisplayLength,
			vis        = settings.fnRecordsDisplay(),
			all        = len === -1;
	
		return str.
			replace(/_START_/g, formatter.call( settings, start ) ).
			replace(/_END_/g,   formatter.call( settings, settings.fnDisplayEnd() ) ).
			replace(/_MAX_/g,   formatter.call( settings, settings.fnRecordsTotal() ) ).
			replace(/_TOTAL_/g, formatter.call( settings, vis ) ).
			replace(/_PAGE_/g,  formatter.call( settings, all ? 1 : Math.ceil( start / len ) ) ).
			replace(/_PAGES_/g, formatter.call( settings, all ? 1 : Math.ceil( vis / len ) ) );
	}
	
	
	
	/**
	 * Draw the table for the first time, adding all required features
	 *  @param {object} settings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnInitialise ( settings )
	{
		var i, iLen, iAjaxStart=settings.iInitDisplayStart;
		var columns = settings.aoColumns, column;
		var features = settings.oFeatures;
		var deferLoading = settings.bDeferLoading; // value modified by the draw
	
		/* Ensure that the table data is fully initialised */
		if ( ! settings.bInitialised ) {
			setTimeout( function(){ _fnInitialise( settings ); }, 200 );
			return;
		}
	
		/* Show the display HTML options */
		_fnAddOptionsHtml( settings );
	
		/* Build and draw the header / footer for the table */
		_fnBuildHead( settings );
		_fnDrawHead( settings, settings.aoHeader );
		_fnDrawHead( settings, settings.aoFooter );
	
		/* Okay to show that something is going on now */
		_fnProcessingDisplay( settings, true );
	
		/* Calculate sizes for columns */
		if ( features.bAutoWidth ) {
			_fnCalculateColumnWidths( settings );
		}
	
		for ( i=0, iLen=columns.length ; i<iLen ; i++ ) {
			column = columns[i];
	
			if ( column.sWidth ) {
				column.nTh.style.width = _fnStringToCss( column.sWidth );
			}
		}
	
		_fnCallbackFire( settings, null, 'preInit', [settings] );
	
		// If there is default sorting required - let's do it. The sort function
		// will do the drawing for us. Otherwise we draw the table regardless of the
		// Ajax source - this allows the table to look initialised for Ajax sourcing
		// data (show 'loading' message possibly)
		_fnReDraw( settings );
	
		// Server-side processing init complete is done by _fnAjaxUpdateDraw
		var dataSrc = _fnDataSource( settings );
		if ( dataSrc != 'ssp' || deferLoading ) {
			// if there is an ajax source load the data
			if ( dataSrc == 'ajax' ) {
				_fnBuildAjax( settings, [], function(json) {
					var aData = _fnAjaxDataSrc( settings, json );
	
					// Got the data - add it to the table
					for ( i=0 ; i<aData.length ; i++ ) {
						_fnAddData( settings, aData[i] );
					}
	
					// Reset the init display for cookie saving. We've already done
					// a filter, and therefore cleared it before. So we need to make
					// it appear 'fresh'
					settings.iInitDisplayStart = iAjaxStart;
	
					_fnReDraw( settings );
	
					_fnProcessingDisplay( settings, false );
					_fnInitComplete( settings, json );
				}, settings );
			}
			else {
				_fnProcessingDisplay( settings, false );
				_fnInitComplete( settings );
			}
		}
	}
	
	
	/**
	 * Draw the table for the first time, adding all required features
	 *  @param {object} oSettings dataTables settings object
	 *  @param {object} [json] JSON from the server that completed the table, if using Ajax source
	 *    with client-side processing (optional)
	 *  @memberof DataTable#oApi
	 */
	function _fnInitComplete ( settings, json )
	{
		settings._bInitComplete = true;
	
		// When data was added after the initialisation (data or Ajax) we need to
		// calculate the column sizing
		if ( json || settings.oInit.aaData ) {
			_fnAdjustColumnSizing( settings );
		}
	
		_fnCallbackFire( settings, null, 'plugin-init', [settings, json] );
		_fnCallbackFire( settings, 'aoInitComplete', 'init', [settings, json] );
	}
	
	
	function _fnLengthChange ( settings, val )
	{
		var len = parseInt( val, 10 );
		settings._iDisplayLength = len;
	
		_fnLengthOverflow( settings );
	
		// Fire length change event
		_fnCallbackFire( settings, null, 'length', [settings, len] );
	}
	
	
	/**
	 * Generate the node required for user display length changing
	 *  @param {object} settings dataTables settings object
	 *  @returns {node} Display length feature node
	 *  @memberof DataTable#oApi
	 */
	function _fnFeatureHtmlLength ( settings )
	{
		var
			classes  = settings.oClasses,
			tableId  = settings.sTableId,
			menu     = settings.aLengthMenu,
			d2       = Array.isArray( menu[0] ),
			lengths  = d2 ? menu[0] : menu,
			language = d2 ? menu[1] : menu;
	
		var select = $('<select/>', {
			'name':          tableId+'_length',
			'aria-controls': tableId,
			'class':         classes.sLengthSelect
		} );
	
		for ( var i=0, ien=lengths.length ; i<ien ; i++ ) {
			select[0][ i ] = new Option(
				typeof language[i] === 'number' ?
					settings.fnFormatNumber( language[i] ) :
					language[i],
				lengths[i]
			);
		}
	
		var div = $('<div><label/></div>').addClass( classes.sLength );
		if ( ! settings.aanFeatures.l ) {
			div[0].id = tableId+'_length';
		}
	
		div.children().append(
			settings.oLanguage.sLengthMenu.replace( '_MENU_', select[0].outerHTML )
		);
	
		// Can't use `select` variable as user might provide their own and the
		// reference is broken by the use of outerHTML
		$('select', div)
			.val( settings._iDisplayLength )
			.on( 'change.DT', function(e) {
				_fnLengthChange( settings, $(this).val() );
				_fnDraw( settings );
			} );
	
		// Update node value whenever anything changes the table's length
		$(settings.nTable).on( 'length.dt.DT', function (e, s, len) {
			if ( settings === s ) {
				$('select', div).val( len );
			}
		} );
	
		return div[0];
	}
	
	
	
	/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
	 * Note that most of the paging logic is done in
	 * DataTable.ext.pager
	 */
	
	/**
	 * Generate the node required for default pagination
	 *  @param {object} oSettings dataTables settings object
	 *  @returns {node} Pagination feature node
	 *  @memberof DataTable#oApi
	 */
	function _fnFeatureHtmlPaginate ( settings )
	{
		var
			type   = settings.sPaginationType,
			plugin = DataTable.ext.pager[ type ],
			modern = typeof plugin === 'function',
			redraw = function( settings ) {
				_fnDraw( settings );
			},
			node = $('<div/>').addClass( settings.oClasses.sPaging + type )[0],
			features = settings.aanFeatures;
	
		if ( ! modern ) {
			plugin.fnInit( settings, node, redraw );
		}
	
		/* Add a draw callback for the pagination on first instance, to update the paging display */
		if ( ! features.p )
		{
			node.id = settings.sTableId+'_paginate';
	
			settings.aoDrawCallback.push( {
				"fn": function( settings ) {
					if ( modern ) {
						var
							start      = settings._iDisplayStart,
							len        = settings._iDisplayLength,
							visRecords = settings.fnRecordsDisplay(),
							all        = len === -1,
							page = all ? 0 : Math.ceil( start / len ),
							pages = all ? 1 : Math.ceil( visRecords / len ),
							buttons = plugin(page, pages),
							i, ien;
	
						for ( i=0, ien=features.p.length ; i<ien ; i++ ) {
							_fnRenderer( settings, 'pageButton' )(
								settings, features.p[i], i, buttons, page, pages
							);
						}
					}
					else {
						plugin.fnUpdate( settings, redraw );
					}
				},
				"sName": "pagination"
			} );
		}
	
		return node;
	}
	
	
	/**
	 * Alter the display settings to change the page
	 *  @param {object} settings DataTables settings object
	 *  @param {string|int} action Paging action to take: "first", "previous",
	 *    "next" or "last" or page number to jump to (integer)
	 *  @param [bool] redraw Automatically draw the update or not
	 *  @returns {bool} true page has changed, false - no change
	 *  @memberof DataTable#oApi
	 */
	function _fnPageChange ( settings, action, redraw )
	{
		var
			start     = settings._iDisplayStart,
			len       = settings._iDisplayLength,
			records   = settings.fnRecordsDisplay();
	
		if ( records === 0 || len === -1 )
		{
			start = 0;
		}
		else if ( typeof action === "number" )
		{
			start = action * len;
	
			if ( start > records )
			{
				start = 0;
			}
		}
		else if ( action == "first" )
		{
			start = 0;
		}
		else if ( action == "previous" )
		{
			start = len >= 0 ?
				start - len :
				0;
	
			if ( start < 0 )
			{
			  start = 0;
			}
		}
		else if ( action == "next" )
		{
			if ( start + len < records )
			{
				start += len;
			}
		}
		else if ( action == "last" )
		{
			start = Math.floor( (records-1) / len) * len;
		}
		else
		{
			_fnLog( settings, 0, "Unknown paging action: "+action, 5 );
		}
	
		var changed = settings._iDisplayStart !== start;
		settings._iDisplayStart = start;
	
		if ( changed ) {
			_fnCallbackFire( settings, null, 'page', [settings] );
	
			if ( redraw ) {
				_fnDraw( settings );
			}
		}
		else {
			// No change event - paging was called, but no change
			_fnCallbackFire( settings, null, 'page-nc', [settings] );
		}
	
		return changed;
	}
	
	
	
	/**
	 * Generate the node required for the processing node
	 *  @param {object} settings dataTables settings object
	 *  @returns {node} Processing element
	 *  @memberof DataTable#oApi
	 */
	function _fnFeatureHtmlProcessing ( settings )
	{
		return $('<div/>', {
				'id': ! settings.aanFeatures.r ? settings.sTableId+'_processing' : null,
				'class': settings.oClasses.sProcessing,
				'role': 'status'
			} )
			.html( settings.oLanguage.sProcessing )
			.append('<div><div></div><div></div><div></div><div></div></div>')
			.insertBefore( settings.nTable )[0];
	}
	
	
	/**
	 * Display or hide the processing indicator
	 *  @param {object} settings dataTables settings object
	 *  @param {bool} show Show the processing indicator (true) or not (false)
	 *  @memberof DataTable#oApi
	 */
	function _fnProcessingDisplay ( settings, show )
	{
		if ( settings.oFeatures.bProcessing ) {
			$(settings.aanFeatures.r).css( 'display', show ? 'block' : 'none' );
		}
	
		_fnCallbackFire( settings, null, 'processing', [settings, show] );
	}
	
	/**
	 * Add any control elements for the table - specifically scrolling
	 *  @param {object} settings dataTables settings object
	 *  @returns {node} Node to add to the DOM
	 *  @memberof DataTable#oApi
	 */
	function _fnFeatureHtmlTable ( settings )
	{
		var table = $(settings.nTable);
	
		// Scrolling from here on in
		var scroll = settings.oScroll;
	
		if ( scroll.sX === '' && scroll.sY === '' ) {
			return settings.nTable;
		}
	
		var scrollX = scroll.sX;
		var scrollY = scroll.sY;
		var classes = settings.oClasses;
		var caption = table.children('caption');
		var captionSide = caption.length ? caption[0]._captionSide : null;
		var headerClone = $( table[0].cloneNode(false) );
		var footerClone = $( table[0].cloneNode(false) );
		var footer = table.children('tfoot');
		var _div = '<div/>';
		var size = function ( s ) {
			return !s ? null : _fnStringToCss( s );
		};
	
		if ( ! footer.length ) {
			footer = null;
		}
	
		/*
		 * The HTML structure that we want to generate in this function is:
		 *  div - scroller
		 *    div - scroll head
		 *      div - scroll head inner
		 *        table - scroll head table
		 *          thead - thead
		 *    div - scroll body
		 *      table - table (master table)
		 *        thead - thead clone for sizing
		 *        tbody - tbody
		 *    div - scroll foot
		 *      div - scroll foot inner
		 *        table - scroll foot table
		 *          tfoot - tfoot
		 */
		var scroller = $( _div, { 'class': classes.sScrollWrapper } )
			.append(
				$(_div, { 'class': classes.sScrollHead } )
					.css( {
						overflow: 'hidden',
						position: 'relative',
						border: 0,
						width: scrollX ? size(scrollX) : '100%'
					} )
					.append(
						$(_div, { 'class': classes.sScrollHeadInner } )
							.css( {
								'box-sizing': 'content-box',
								width: scroll.sXInner || '100%'
							} )
							.append(
								headerClone
									.removeAttr('id')
									.css( 'margin-left', 0 )
									.append( captionSide === 'top' ? caption : null )
									.append(
										table.children('thead')
									)
							)
					)
			)
			.append(
				$(_div, { 'class': classes.sScrollBody } )
					.css( {
						position: 'relative',
						overflow: 'auto',
						width: size( scrollX )
					} )
					.append( table )
			);
	
		if ( footer ) {
			scroller.append(
				$(_div, { 'class': classes.sScrollFoot } )
					.css( {
						overflow: 'hidden',
						border: 0,
						width: scrollX ? size(scrollX) : '100%'
					} )
					.append(
						$(_div, { 'class': classes.sScrollFootInner } )
							.append(
								footerClone
									.removeAttr('id')
									.css( 'margin-left', 0 )
									.append( captionSide === 'bottom' ? caption : null )
									.append(
										table.children('tfoot')
									)
							)
					)
			);
		}
	
		var children = scroller.children();
		var scrollHead = children[0];
		var scrollBody = children[1];
		var scrollFoot = footer ? children[2] : null;
	
		// When the body is scrolled, then we also want to scroll the headers
		if ( scrollX ) {
			$(scrollBody).on( 'scroll.DT', function (e) {
				var scrollLeft = this.scrollLeft;
	
				scrollHead.scrollLeft = scrollLeft;
	
				if ( footer ) {
					scrollFoot.scrollLeft = scrollLeft;
				}
			} );
		}
	
		$(scrollBody).css('max-height', scrollY);
		if (! scroll.bCollapse) {
			$(scrollBody).css('height', scrollY);
		}
	
		settings.nScrollHead = scrollHead;
		settings.nScrollBody = scrollBody;
		settings.nScrollFoot = scrollFoot;
	
		// On redraw - align columns
		settings.aoDrawCallback.push( {
			"fn": _fnScrollDraw,
			"sName": "scrolling"
		} );
	
		return scroller[0];
	}
	
	
	
	/**
	 * Update the header, footer and body tables for resizing - i.e. column
	 * alignment.
	 *
	 * Welcome to the most horrible function DataTables. The process that this
	 * function follows is basically:
	 *   1. Re-create the table inside the scrolling div
	 *   2. Take live measurements from the DOM
	 *   3. Apply the measurements to align the columns
	 *   4. Clean up
	 *
	 *  @param {object} settings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnScrollDraw ( settings )
	{
		// Given that this is such a monster function, a lot of variables are use
		// to try and keep the minimised size as small as possible
		var
			scroll         = settings.oScroll,
			scrollX        = scroll.sX,
			scrollXInner   = scroll.sXInner,
			scrollY        = scroll.sY,
			barWidth       = scroll.iBarWidth,
			divHeader      = $(settings.nScrollHead),
			divHeaderStyle = divHeader[0].style,
			divHeaderInner = divHeader.children('div'),
			divHeaderInnerStyle = divHeaderInner[0].style,
			divHeaderTable = divHeaderInner.children('table'),
			divBodyEl      = settings.nScrollBody,
			divBody        = $(divBodyEl),
			divBodyStyle   = divBodyEl.style,
			divFooter      = $(settings.nScrollFoot),
			divFooterInner = divFooter.children('div'),
			divFooterTable = divFooterInner.children('table'),
			header         = $(settings.nTHead),
			table          = $(settings.nTable),
			tableEl        = table[0],
			tableStyle     = tableEl.style,
			footer         = settings.nTFoot ? $(settings.nTFoot) : null,
			browser        = settings.oBrowser,
			ie67           = browser.bScrollOversize,
			dtHeaderCells  = _pluck( settings.aoColumns, 'nTh' ),
			headerTrgEls, footerTrgEls,
			headerSrcEls, footerSrcEls,
			headerCopy, footerCopy,
			headerWidths=[], footerWidths=[],
			headerContent=[], footerContent=[],
			idx, correction, sanityWidth,
			zeroOut = function(nSizer) {
				var style = nSizer.style;
				style.paddingTop = "0";
				style.paddingBottom = "0";
				style.borderTopWidth = "0";
				style.borderBottomWidth = "0";
				style.height = 0;
			};
	
		// If the scrollbar visibility has changed from the last draw, we need to
		// adjust the column sizes as the table width will have changed to account
		// for the scrollbar
		var scrollBarVis = divBodyEl.scrollHeight > divBodyEl.clientHeight;
		
		if ( settings.scrollBarVis !== scrollBarVis && settings.scrollBarVis !== undefined ) {
			settings.scrollBarVis = scrollBarVis;
			_fnAdjustColumnSizing( settings );
			return; // adjust column sizing will call this function again
		}
		else {
			settings.scrollBarVis = scrollBarVis;
		}
	
		/*
		 * 1. Re-create the table inside the scrolling div
		 */
	
		// Remove the old minimised thead and tfoot elements in the inner table
		table.children('thead, tfoot').remove();
	
		if ( footer ) {
			footerCopy = footer.clone().prependTo( table );
			footerTrgEls = footer.find('tr'); // the original tfoot is in its own table and must be sized
			footerSrcEls = footerCopy.find('tr');
			footerCopy.find('[id]').removeAttr('id');
		}
	
		// Clone the current header and footer elements and then place it into the inner table
		headerCopy = header.clone().prependTo( table );
		headerTrgEls = header.find('tr'); // original header is in its own table
		headerSrcEls = headerCopy.find('tr');
		headerCopy.find('th, td').removeAttr('tabindex');
		headerCopy.find('[id]').removeAttr('id');
	
	
		/*
		 * 2. Take live measurements from the DOM - do not alter the DOM itself!
		 */
	
		// Remove old sizing and apply the calculated column widths
		// Get the unique column headers in the newly created (cloned) header. We want to apply the
		// calculated sizes to this header
		if ( ! scrollX )
		{
			divBodyStyle.width = '100%';
			divHeader[0].style.width = '100%';
		}
	
		$.each( _fnGetUniqueThs( settings, headerCopy ), function ( i, el ) {
			idx = _fnVisibleToColumnIndex( settings, i );
			el.style.width = settings.aoColumns[idx].sWidth;
		} );
	
		if ( footer ) {
			_fnApplyToChildren( function(n) {
				n.style.width = "";
			}, footerSrcEls );
		}
	
		// Size the table as a whole
		sanityWidth = table.outerWidth();
		if ( scrollX === "" ) {
			// No x scrolling
			tableStyle.width = "100%";
	
			// IE7 will make the width of the table when 100% include the scrollbar
			// - which is shouldn't. When there is a scrollbar we need to take this
			// into account.
			if ( ie67 && (table.find('tbody').height() > divBodyEl.offsetHeight ||
				divBody.css('overflow-y') == "scroll")
			) {
				tableStyle.width = _fnStringToCss( table.outerWidth() - barWidth);
			}
	
			// Recalculate the sanity width
			sanityWidth = table.outerWidth();
		}
		else if ( scrollXInner !== "" ) {
			// legacy x scroll inner has been given - use it
			tableStyle.width = _fnStringToCss(scrollXInner);
	
			// Recalculate the sanity width
			sanityWidth = table.outerWidth();
		}
	
		// Hidden header should have zero height, so remove padding and borders. Then
		// set the width based on the real headers
	
		// Apply all styles in one pass
		_fnApplyToChildren( zeroOut, headerSrcEls );
	
		// Read all widths in next pass
		_fnApplyToChildren( function(nSizer) {
			var style = window.getComputedStyle ?
				window.getComputedStyle(nSizer).width :
				_fnStringToCss( $(nSizer).width() );
	
			headerContent.push( nSizer.innerHTML );
			headerWidths.push( style );
		}, headerSrcEls );
	
		// Apply all widths in final pass
		_fnApplyToChildren( function(nToSize, i) {
			nToSize.style.width = headerWidths[i];
		}, headerTrgEls );
	
		$(headerSrcEls).css('height', 0);
	
		/* Same again with the footer if we have one */
		if ( footer )
		{
			_fnApplyToChildren( zeroOut, footerSrcEls );
	
			_fnApplyToChildren( function(nSizer) {
				footerContent.push( nSizer.innerHTML );
				footerWidths.push( _fnStringToCss( $(nSizer).css('width') ) );
			}, footerSrcEls );
	
			_fnApplyToChildren( function(nToSize, i) {
				nToSize.style.width = footerWidths[i];
			}, footerTrgEls );
	
			$(footerSrcEls).height(0);
		}
	
	
		/*
		 * 3. Apply the measurements
		 */
	
		// "Hide" the header and footer that we used for the sizing. We need to keep
		// the content of the cell so that the width applied to the header and body
		// both match, but we want to hide it completely. We want to also fix their
		// width to what they currently are
		_fnApplyToChildren( function(nSizer, i) {
			nSizer.innerHTML = '<div class="dataTables_sizing">'+headerContent[i]+'</div>';
			nSizer.childNodes[0].style.height = "0";
			nSizer.childNodes[0].style.overflow = "hidden";
			nSizer.style.width = headerWidths[i];
		}, headerSrcEls );
	
		if ( footer )
		{
			_fnApplyToChildren( function(nSizer, i) {
				nSizer.innerHTML = '<div class="dataTables_sizing">'+footerContent[i]+'</div>';
				nSizer.childNodes[0].style.height = "0";
				nSizer.childNodes[0].style.overflow = "hidden";
				nSizer.style.width = footerWidths[i];
			}, footerSrcEls );
		}
	
		// Sanity check that the table is of a sensible width. If not then we are going to get
		// misalignment - try to prevent this by not allowing the table to shrink below its min width
		if ( Math.round(table.outerWidth()) < Math.round(sanityWidth) )
		{
			// The min width depends upon if we have a vertical scrollbar visible or not */
			correction = ((divBodyEl.scrollHeight > divBodyEl.offsetHeight ||
				divBody.css('overflow-y') == "scroll")) ?
					sanityWidth+barWidth :
					sanityWidth;
	
			// IE6/7 are a law unto themselves...
			if ( ie67 && (divBodyEl.scrollHeight >
				divBodyEl.offsetHeight || divBody.css('overflow-y') == "scroll")
			) {
				tableStyle.width = _fnStringToCss( correction-barWidth );
			}
	
			// And give the user a warning that we've stopped the table getting too small
			if ( scrollX === "" || scrollXInner !== "" ) {
				_fnLog( settings, 1, 'Possible column misalignment', 6 );
			}
		}
		else
		{
			correction = '100%';
		}
	
		// Apply to the container elements
		divBodyStyle.width = _fnStringToCss( correction );
		divHeaderStyle.width = _fnStringToCss( correction );
	
		if ( footer ) {
			settings.nScrollFoot.style.width = _fnStringToCss( correction );
		}
	
	
		/*
		 * 4. Clean up
		 */
		if ( ! scrollY ) {
			/* IE7< puts a vertical scrollbar in place (when it shouldn't be) due to subtracting
			 * the scrollbar height from the visible display, rather than adding it on. We need to
			 * set the height in order to sort this. Don't want to do it in any other browsers.
			 */
			if ( ie67 ) {
				divBodyStyle.height = _fnStringToCss( tableEl.offsetHeight+barWidth );
			}
		}
	
		/* Finally set the width's of the header and footer tables */
		var iOuterWidth = table.outerWidth();
		divHeaderTable[0].style.width = _fnStringToCss( iOuterWidth );
		divHeaderInnerStyle.width = _fnStringToCss( iOuterWidth );
	
		// Figure out if there are scrollbar present - if so then we need a the header and footer to
		// provide a bit more space to allow "overflow" scrolling (i.e. past the scrollbar)
		var bScrolling = table.height() > divBodyEl.clientHeight || divBody.css('overflow-y') == "scroll";
		var padding = 'padding' + (browser.bScrollbarLeft ? 'Left' : 'Right' );
		divHeaderInnerStyle[ padding ] = bScrolling ? barWidth+"px" : "0px";
	
		if ( footer ) {
			divFooterTable[0].style.width = _fnStringToCss( iOuterWidth );
			divFooterInner[0].style.width = _fnStringToCss( iOuterWidth );
			divFooterInner[0].style[padding] = bScrolling ? barWidth+"px" : "0px";
		}
	
		// Correct DOM ordering for colgroup - comes before the thead
		table.children('colgroup').insertBefore( table.children('thead') );
	
		/* Adjust the position of the header in case we loose the y-scrollbar */
		divBody.trigger('scroll');
	
		// If sorting or filtering has occurred, jump the scrolling back to the top
		// only if we aren't holding the position
		if ( (settings.bSorted || settings.bFiltered) && ! settings._drawHold ) {
			divBodyEl.scrollTop = 0;
		}
	}
	
	
	
	/**
	 * Apply a given function to the display child nodes of an element array (typically
	 * TD children of TR rows
	 *  @param {function} fn Method to apply to the objects
	 *  @param array {nodes} an1 List of elements to look through for display children
	 *  @param array {nodes} an2 Another list (identical structure to the first) - optional
	 *  @memberof DataTable#oApi
	 */
	function _fnApplyToChildren( fn, an1, an2 )
	{
		var index=0, i=0, iLen=an1.length;
		var nNode1, nNode2;
	
		while ( i < iLen ) {
			nNode1 = an1[i].firstChild;
			nNode2 = an2 ? an2[i].firstChild : null;
	
			while ( nNode1 ) {
				if ( nNode1.nodeType === 1 ) {
					if ( an2 ) {
						fn( nNode1, nNode2, index );
					}
					else {
						fn( nNode1, index );
					}
	
					index++;
				}
	
				nNode1 = nNode1.nextSibling;
				nNode2 = an2 ? nNode2.nextSibling : null;
			}
	
			i++;
		}
	}
	
	
	
	var __re_html_remove = /<.*?>/g;
	
	
	/**
	 * Calculate the width of columns for the table
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnCalculateColumnWidths ( oSettings )
	{
		var
			table = oSettings.nTable,
			columns = oSettings.aoColumns,
			scroll = oSettings.oScroll,
			scrollY = scroll.sY,
			scrollX = scroll.sX,
			scrollXInner = scroll.sXInner,
			columnCount = columns.length,
			visibleColumns = _fnGetColumns( oSettings, 'bVisible' ),
			headerCells = $('th', oSettings.nTHead),
			tableWidthAttr = table.getAttribute('width'), // from DOM element
			tableContainer = table.parentNode,
			userInputs = false,
			i, column, columnIdx, width, outerWidth,
			browser = oSettings.oBrowser,
			ie67 = browser.bScrollOversize;
	
		var styleWidth = table.style.width;
		if ( styleWidth && styleWidth.indexOf('%') !== -1 ) {
			tableWidthAttr = styleWidth;
		}
	
		/* Convert any user input sizes into pixel sizes */
		for ( i=0 ; i<visibleColumns.length ; i++ ) {
			column = columns[ visibleColumns[i] ];
	
			if ( column.sWidth !== null ) {
				column.sWidth = _fnConvertToWidth( column.sWidthOrig, tableContainer );
	
				userInputs = true;
			}
		}
	
		/* If the number of columns in the DOM equals the number that we have to
		 * process in DataTables, then we can use the offsets that are created by
		 * the web- browser. No custom sizes can be set in order for this to happen,
		 * nor scrolling used
		 */
		if ( ie67 || ! userInputs && ! scrollX && ! scrollY &&
		     columnCount == _fnVisbleColumns( oSettings ) &&
		     columnCount == headerCells.length
		) {
			for ( i=0 ; i<columnCount ; i++ ) {
				var colIdx = _fnVisibleToColumnIndex( oSettings, i );
	
				if ( colIdx !== null ) {
					columns[ colIdx ].sWidth = _fnStringToCss( headerCells.eq(i).width() );
				}
			}
		}
		else
		{
			// Otherwise construct a single row, worst case, table with the widest
			// node in the data, assign any user defined widths, then insert it into
			// the DOM and allow the browser to do all the hard work of calculating
			// table widths
			var tmpTable = $(table).clone() // don't use cloneNode - IE8 will remove events on the main table
				.css( 'visibility', 'hidden' )
				.removeAttr( 'id' );
	
			// Clean up the table body
			tmpTable.find('tbody tr').remove();
			var tr = $('<tr/>').appendTo( tmpTable.find('tbody') );
	
			// Clone the table header and footer - we can't use the header / footer
			// from the cloned table, since if scrolling is active, the table's
			// real header and footer are contained in different table tags
			tmpTable.find('thead, tfoot').remove();
			tmpTable
				.append( $(oSettings.nTHead).clone() )
				.append( $(oSettings.nTFoot).clone() );
	
			// Remove any assigned widths from the footer (from scrolling)
			tmpTable.find('tfoot th, tfoot td').css('width', '');
	
			// Apply custom sizing to the cloned header
			headerCells = _fnGetUniqueThs( oSettings, tmpTable.find('thead')[0] );
	
			for ( i=0 ; i<visibleColumns.length ; i++ ) {
				column = columns[ visibleColumns[i] ];
	
				headerCells[i].style.width = column.sWidthOrig !== null && column.sWidthOrig !== '' ?
					_fnStringToCss( column.sWidthOrig ) :
					'';
	
				// For scrollX we need to force the column width otherwise the
				// browser will collapse it. If this width is smaller than the
				// width the column requires, then it will have no effect
				if ( column.sWidthOrig && scrollX ) {
					$( headerCells[i] ).append( $('<div/>').css( {
						width: column.sWidthOrig,
						margin: 0,
						padding: 0,
						border: 0,
						height: 1
					} ) );
				}
			}
	
			// Find the widest cell for each column and put it into the table
			if ( oSettings.aoData.length ) {
				for ( i=0 ; i<visibleColumns.length ; i++ ) {
					columnIdx = visibleColumns[i];
					column = columns[ columnIdx ];
	
					$( _fnGetWidestNode( oSettings, columnIdx ) )
						.clone( false )
						.append( column.sContentPadding )
						.appendTo( tr );
				}
			}
	
			// Tidy the temporary table - remove name attributes so there aren't
			// duplicated in the dom (radio elements for example)
			$('[name]', tmpTable).removeAttr('name');
	
			// Table has been built, attach to the document so we can work with it.
			// A holding element is used, positioned at the top of the container
			// with minimal height, so it has no effect on if the container scrolls
			// or not. Otherwise it might trigger scrolling when it actually isn't
			// needed
			var holder = $('<div/>').css( scrollX || scrollY ?
					{
						position: 'absolute',
						top: 0,
						left: 0,
						height: 1,
						right: 0,
						overflow: 'hidden'
					} :
					{}
				)
				.append( tmpTable )
				.appendTo( tableContainer );
	
			// When scrolling (X or Y) we want to set the width of the table as 
			// appropriate. However, when not scrolling leave the table width as it
			// is. This results in slightly different, but I think correct behaviour
			if ( scrollX && scrollXInner ) {
				tmpTable.width( scrollXInner );
			}
			else if ( scrollX ) {
				tmpTable.css( 'width', 'auto' );
				tmpTable.removeAttr('width');
	
				// If there is no width attribute or style, then allow the table to
				// collapse
				if ( tmpTable.width() < tableContainer.clientWidth && tableWidthAttr ) {
					tmpTable.width( tableContainer.clientWidth );
				}
			}
			else if ( scrollY ) {
				tmpTable.width( tableContainer.clientWidth );
			}
			else if ( tableWidthAttr ) {
				tmpTable.width( tableWidthAttr );
			}
	
			// Get the width of each column in the constructed table - we need to
			// know the inner width (so it can be assigned to the other table's
			// cells) and the outer width so we can calculate the full width of the
			// table. This is safe since DataTables requires a unique cell for each
			// column, but if ever a header can span multiple columns, this will
			// need to be modified.
			var total = 0;
			for ( i=0 ; i<visibleColumns.length ; i++ ) {
				var cell = $(headerCells[i]);
				var border = cell.outerWidth() - cell.width();
	
				// Use getBounding... where possible (not IE8-) because it can give
				// sub-pixel accuracy, which we then want to round up!
				var bounding = browser.bBounding ?
					Math.ceil( headerCells[i].getBoundingClientRect().width ) :
					cell.outerWidth();
	
				// Total is tracked to remove any sub-pixel errors as the outerWidth
				// of the table might not equal the total given here (IE!).
				total += bounding;
	
				// Width for each column to use
				columns[ visibleColumns[i] ].sWidth = _fnStringToCss( bounding - border );
			}
	
			table.style.width = _fnStringToCss( total );
	
			// Finished with the table - ditch it
			holder.remove();
		}
	
		// If there is a width attr, we want to attach an event listener which
		// allows the table sizing to automatically adjust when the window is
		// resized. Use the width attr rather than CSS, since we can't know if the
		// CSS is a relative value or absolute - DOM read is always px.
		if ( tableWidthAttr ) {
			table.style.width = _fnStringToCss( tableWidthAttr );
		}
	
		if ( (tableWidthAttr || scrollX) && ! oSettings._reszEvt ) {
			var bindResize = function () {
				$(window).on('resize.DT-'+oSettings.sInstance, _fnThrottle( function () {
					_fnAdjustColumnSizing( oSettings );
				} ) );
			};
	
			// IE6/7 will crash if we bind a resize event handler on page load.
			// To be removed in 1.11 which drops IE6/7 support
			if ( ie67 ) {
				setTimeout( bindResize, 1000 );
			}
			else {
				bindResize();
			}
	
			oSettings._reszEvt = true;
		}
	}
	
	
	/**
	 * Throttle the calls to a function. Arguments and context are maintained for
	 * the throttled function
	 *  @param {function} fn Function to be called
	 *  @param {int} [freq=200] call frequency in mS
	 *  @returns {function} wrapped function
	 *  @memberof DataTable#oApi
	 */
	var _fnThrottle = DataTable.util.throttle;
	
	
	/**
	 * Convert a CSS unit width to pixels (e.g. 2em)
	 *  @param {string} width width to be converted
	 *  @param {node} parent parent to get the with for (required for relative widths) - optional
	 *  @returns {int} width in pixels
	 *  @memberof DataTable#oApi
	 */
	function _fnConvertToWidth ( width, parent )
	{
		if ( ! width ) {
			return 0;
		}
	
		var n = $('<div/>')
			.css( 'width', _fnStringToCss( width ) )
			.appendTo( parent || document.body );
	
		var val = n[0].offsetWidth;
		n.remove();
	
		return val;
	}
	
	
	/**
	 * Get the widest node
	 *  @param {object} settings dataTables settings object
	 *  @param {int} colIdx column of interest
	 *  @returns {node} widest table node
	 *  @memberof DataTable#oApi
	 */
	function _fnGetWidestNode( settings, colIdx )
	{
		var idx = _fnGetMaxLenString( settings, colIdx );
		if ( idx < 0 ) {
			return null;
		}
	
		var data = settings.aoData[ idx ];
		return ! data.nTr ? // Might not have been created when deferred rendering
			$('<td/>').html( _fnGetCellData( settings, idx, colIdx, 'display' ) )[0] :
			data.anCells[ colIdx ];
	}
	
	
	/**
	 * Get the maximum strlen for each data column
	 *  @param {object} settings dataTables settings object
	 *  @param {int} colIdx column of interest
	 *  @returns {string} max string length for each column
	 *  @memberof DataTable#oApi
	 */
	function _fnGetMaxLenString( settings, colIdx )
	{
		var s, max=-1, maxIdx = -1;
	
		for ( var i=0, ien=settings.aoData.length ; i<ien ; i++ ) {
			s = _fnGetCellData( settings, i, colIdx, 'display' )+'';
			s = s.replace( __re_html_remove, '' );
			s = s.replace( /&nbsp;/g, ' ' );
	
			if ( s.length > max ) {
				max = s.length;
				maxIdx = i;
			}
		}
	
		return maxIdx;
	}
	
	
	/**
	 * Append a CSS unit (only if required) to a string
	 *  @param {string} value to css-ify
	 *  @returns {string} value with css unit
	 *  @memberof DataTable#oApi
	 */
	function _fnStringToCss( s )
	{
		if ( s === null ) {
			return '0px';
		}
	
		if ( typeof s == 'number' ) {
			return s < 0 ?
				'0px' :
				s+'px';
		}
	
		// Check it has a unit character already
		return s.match(/\d$/) ?
			s+'px' :
			s;
	}
	
	
	
	function _fnSortFlatten ( settings )
	{
		var
			i, iLen, k, kLen,
			aSort = [],
			aiOrig = [],
			aoColumns = settings.aoColumns,
			aDataSort, iCol, sType, srcCol,
			fixed = settings.aaSortingFixed,
			fixedObj = $.isPlainObject( fixed ),
			nestedSort = [],
			add = function ( a ) {
				if ( a.length && ! Array.isArray( a[0] ) ) {
					// 1D array
					nestedSort.push( a );
				}
				else {
					// 2D array
					$.merge( nestedSort, a );
				}
			};
	
		// Build the sort array, with pre-fix and post-fix options if they have been
		// specified
		if ( Array.isArray( fixed ) ) {
			add( fixed );
		}
	
		if ( fixedObj && fixed.pre ) {
			add( fixed.pre );
		}
	
		add( settings.aaSorting );
	
		if (fixedObj && fixed.post ) {
			add( fixed.post );
		}
	
		for ( i=0 ; i<nestedSort.length ; i++ )
		{
			srcCol = nestedSort[i][0];
			aDataSort = aoColumns[ srcCol ].aDataSort;
	
			for ( k=0, kLen=aDataSort.length ; k<kLen ; k++ )
			{
				iCol = aDataSort[k];
				sType = aoColumns[ iCol ].sType || 'string';
	
				if ( nestedSort[i]._idx === undefined ) {
					nestedSort[i]._idx = $.inArray( nestedSort[i][1], aoColumns[iCol].asSorting );
				}
	
				aSort.push( {
					src:       srcCol,
					col:       iCol,
					dir:       nestedSort[i][1],
					index:     nestedSort[i]._idx,
					type:      sType,
					formatter: DataTable.ext.type.order[ sType+"-pre" ]
				} );
			}
		}
	
		return aSort;
	}
	
	/**
	 * Change the order of the table
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 *  @todo This really needs split up!
	 */
	function _fnSort ( oSettings )
	{
		var
			i, ien, iLen, j, jLen, k, kLen,
			sDataType, nTh,
			aiOrig = [],
			oExtSort = DataTable.ext.type.order,
			aoData = oSettings.aoData,
			aoColumns = oSettings.aoColumns,
			aDataSort, data, iCol, sType, oSort,
			formatters = 0,
			sortCol,
			displayMaster = oSettings.aiDisplayMaster,
			aSort;
	
		// Resolve any column types that are unknown due to addition or invalidation
		// @todo Can this be moved into a 'data-ready' handler which is called when
		//   data is going to be used in the table?
		_fnColumnTypes( oSettings );
	
		aSort = _fnSortFlatten( oSettings );
	
		for ( i=0, ien=aSort.length ; i<ien ; i++ ) {
			sortCol = aSort[i];
	
			// Track if we can use the fast sort algorithm
			if ( sortCol.formatter ) {
				formatters++;
			}
	
			// Load the data needed for the sort, for each cell
			_fnSortData( oSettings, sortCol.col );
		}
	
		/* No sorting required if server-side or no sorting array */
		if ( _fnDataSource( oSettings ) != 'ssp' && aSort.length !== 0 )
		{
			// Create a value - key array of the current row positions such that we can use their
			// current position during the sort, if values match, in order to perform stable sorting
			for ( i=0, iLen=displayMaster.length ; i<iLen ; i++ ) {
				aiOrig[ displayMaster[i] ] = i;
			}
	
			/* Do the sort - here we want multi-column sorting based on a given data source (column)
			 * and sorting function (from oSort) in a certain direction. It's reasonably complex to
			 * follow on it's own, but this is what we want (example two column sorting):
			 *  fnLocalSorting = function(a,b){
			 *    var iTest;
			 *    iTest = oSort['string-asc']('data11', 'data12');
			 *      if (iTest !== 0)
			 *        return iTest;
			 *    iTest = oSort['numeric-desc']('data21', 'data22');
			 *    if (iTest !== 0)
			 *      return iTest;
			 *    return oSort['numeric-asc']( aiOrig[a], aiOrig[b] );
			 *  }
			 * Basically we have a test for each sorting column, if the data in that column is equal,
			 * test the next column. If all columns match, then we use a numeric sort on the row
			 * positions in the original data array to provide a stable sort.
			 *
			 * Note - I know it seems excessive to have two sorting methods, but the first is around
			 * 15% faster, so the second is only maintained for backwards compatibility with sorting
			 * methods which do not have a pre-sort formatting function.
			 */
			if ( formatters === aSort.length ) {
				// All sort types have formatting functions
				displayMaster.sort( function ( a, b ) {
					var
						x, y, k, test, sort,
						len=aSort.length,
						dataA = aoData[a]._aSortData,
						dataB = aoData[b]._aSortData;
	
					for ( k=0 ; k<len ; k++ ) {
						sort = aSort[k];
	
						x = dataA[ sort.col ];
						y = dataB[ sort.col ];
	
						test = x<y ? -1 : x>y ? 1 : 0;
						if ( test !== 0 ) {
							return sort.dir === 'asc' ? test : -test;
						}
					}
	
					x = aiOrig[a];
					y = aiOrig[b];
					return x<y ? -1 : x>y ? 1 : 0;
				} );
			}
			else {
				// Depreciated - remove in 1.11 (providing a plug-in option)
				// Not all sort types have formatting methods, so we have to call their sorting
				// methods.
				displayMaster.sort( function ( a, b ) {
					var
						x, y, k, l, test, sort, fn,
						len=aSort.length,
						dataA = aoData[a]._aSortData,
						dataB = aoData[b]._aSortData;
	
					for ( k=0 ; k<len ; k++ ) {
						sort = aSort[k];
	
						x = dataA[ sort.col ];
						y = dataB[ sort.col ];
	
						fn = oExtSort[ sort.type+"-"+sort.dir ] || oExtSort[ "string-"+sort.dir ];
						test = fn( x, y );
						if ( test !== 0 ) {
							return test;
						}
					}
	
					x = aiOrig[a];
					y = aiOrig[b];
					return x<y ? -1 : x>y ? 1 : 0;
				} );
			}
		}
	
		/* Tell the draw function that we have sorted the data */
		oSettings.bSorted = true;
	}
	
	
	function _fnSortAria ( settings )
	{
		var label;
		var nextSort;
		var columns = settings.aoColumns;
		var aSort = _fnSortFlatten( settings );
		var oAria = settings.oLanguage.oAria;
	
		// ARIA attributes - need to loop all columns, to update all (removing old
		// attributes as needed)
		for ( var i=0, iLen=columns.length ; i<iLen ; i++ )
		{
			var col = columns[i];
			var asSorting = col.asSorting;
			var sTitle = col.ariaTitle || col.sTitle.replace( /<.*?>/g, "" );
			var th = col.nTh;
	
			// IE7 is throwing an error when setting these properties with jQuery's
			// attr() and removeAttr() methods...
			th.removeAttribute('aria-sort');
	
			/* In ARIA only the first sorting column can be marked as sorting - no multi-sort option */
			if ( col.bSortable ) {
				if ( aSort.length > 0 && aSort[0].col == i ) {
					th.setAttribute('aria-sort', aSort[0].dir=="asc" ? "ascending" : "descending" );
					nextSort = asSorting[ aSort[0].index+1 ] || asSorting[0];
				}
				else {
					nextSort = asSorting[0];
				}
	
				label = sTitle + ( nextSort === "asc" ?
					oAria.sSortAscending :
					oAria.sSortDescending
				);
			}
			else {
				label = sTitle;
			}
	
			th.setAttribute('aria-label', label);
		}
	}
	
	
	/**
	 * Function to run on user sort request
	 *  @param {object} settings dataTables settings object
	 *  @param {node} attachTo node to attach the handler to
	 *  @param {int} colIdx column sorting index
	 *  @param {boolean} [append=false] Append the requested sort to the existing
	 *    sort if true (i.e. multi-column sort)
	 *  @param {function} [callback] callback function
	 *  @memberof DataTable#oApi
	 */
	function _fnSortListener ( settings, colIdx, append, callback )
	{
		var col = settings.aoColumns[ colIdx ];
		var sorting = settings.aaSorting;
		var asSorting = col.asSorting;
		var nextSortIdx;
		var next = function ( a, overflow ) {
			var idx = a._idx;
			if ( idx === undefined ) {
				idx = $.inArray( a[1], asSorting );
			}
	
			return idx+1 < asSorting.length ?
				idx+1 :
				overflow ?
					null :
					0;
		};
	
		// Convert to 2D array if needed
		if ( typeof sorting[0] === 'number' ) {
			sorting = settings.aaSorting = [ sorting ];
		}
	
		// If appending the sort then we are multi-column sorting
		if ( append && settings.oFeatures.bSortMulti ) {
			// Are we already doing some kind of sort on this column?
			var sortIdx = $.inArray( colIdx, _pluck(sorting, '0') );
	
			if ( sortIdx !== -1 ) {
				// Yes, modify the sort
				nextSortIdx = next( sorting[sortIdx], true );
	
				if ( nextSortIdx === null && sorting.length === 1 ) {
					nextSortIdx = 0; // can't remove sorting completely
				}
	
				if ( nextSortIdx === null ) {
					sorting.splice( sortIdx, 1 );
				}
				else {
					sorting[sortIdx][1] = asSorting[ nextSortIdx ];
					sorting[sortIdx]._idx = nextSortIdx;
				}
			}
			else {
				// No sort on this column yet
				sorting.push( [ colIdx, asSorting[0], 0 ] );
				sorting[sorting.length-1]._idx = 0;
			}
		}
		else if ( sorting.length && sorting[0][0] == colIdx ) {
			// Single column - already sorting on this column, modify the sort
			nextSortIdx = next( sorting[0] );
	
			sorting.length = 1;
			sorting[0][1] = asSorting[ nextSortIdx ];
			sorting[0]._idx = nextSortIdx;
		}
		else {
			// Single column - sort only on this column
			sorting.length = 0;
			sorting.push( [ colIdx, asSorting[0] ] );
			sorting[0]._idx = 0;
		}
	
		// Run the sort by calling a full redraw
		_fnReDraw( settings );
	
		// callback used for async user interaction
		if ( typeof callback == 'function' ) {
			callback( settings );
		}
	}
	
	
	/**
	 * Attach a sort handler (click) to a node
	 *  @param {object} settings dataTables settings object
	 *  @param {node} attachTo node to attach the handler to
	 *  @param {int} colIdx column sorting index
	 *  @param {function} [callback] callback function
	 *  @memberof DataTable#oApi
	 */
	function _fnSortAttachListener ( settings, attachTo, colIdx, callback )
	{
		var col = settings.aoColumns[ colIdx ];
	
		_fnBindAction( attachTo, {}, function (e) {
			/* If the column is not sortable - don't to anything */
			if ( col.bSortable === false ) {
				return;
			}
	
			// If processing is enabled use a timeout to allow the processing
			// display to be shown - otherwise to it synchronously
			if ( settings.oFeatures.bProcessing ) {
				_fnProcessingDisplay( settings, true );
	
				setTimeout( function() {
					_fnSortListener( settings, colIdx, e.shiftKey, callback );
	
					// In server-side processing, the draw callback will remove the
					// processing display
					if ( _fnDataSource( settings ) !== 'ssp' ) {
						_fnProcessingDisplay( settings, false );
					}
				}, 0 );
			}
			else {
				_fnSortListener( settings, colIdx, e.shiftKey, callback );
			}
		} );
	}
	
	
	/**
	 * Set the sorting classes on table's body, Note: it is safe to call this function
	 * when bSort and bSortClasses are false
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnSortingClasses( settings )
	{
		var oldSort = settings.aLastSort;
		var sortClass = settings.oClasses.sSortColumn;
		var sort = _fnSortFlatten( settings );
		var features = settings.oFeatures;
		var i, ien, colIdx;
	
		if ( features.bSort && features.bSortClasses ) {
			// Remove old sorting classes
			for ( i=0, ien=oldSort.length ; i<ien ; i++ ) {
				colIdx = oldSort[i].src;
	
				// Remove column sorting
				$( _pluck( settings.aoData, 'anCells', colIdx ) )
					.removeClass( sortClass + (i<2 ? i+1 : 3) );
			}
	
			// Add new column sorting
			for ( i=0, ien=sort.length ; i<ien ; i++ ) {
				colIdx = sort[i].src;
	
				$( _pluck( settings.aoData, 'anCells', colIdx ) )
					.addClass( sortClass + (i<2 ? i+1 : 3) );
			}
		}
	
		settings.aLastSort = sort;
	}
	
	
	// Get the data to sort a column, be it from cache, fresh (populating the
	// cache), or from a sort formatter
	function _fnSortData( settings, idx )
	{
		// Custom sorting function - provided by the sort data type
		var column = settings.aoColumns[ idx ];
		var customSort = DataTable.ext.order[ column.sSortDataType ];
		var customData;
	
		if ( customSort ) {
			customData = customSort.call( settings.oInstance, settings, idx,
				_fnColumnIndexToVisible( settings, idx )
			);
		}
	
		// Use / populate cache
		var row, cellData;
		var formatter = DataTable.ext.type.order[ column.sType+"-pre" ];
	
		for ( var i=0, ien=settings.aoData.length ; i<ien ; i++ ) {
			row = settings.aoData[i];
	
			if ( ! row._aSortData ) {
				row._aSortData = [];
			}
	
			if ( ! row._aSortData[idx] || customSort ) {
				cellData = customSort ?
					customData[i] : // If there was a custom sort function, use data from there
					_fnGetCellData( settings, i, idx, 'sort' );
	
				row._aSortData[ idx ] = formatter ?
					formatter( cellData ) :
					cellData;
			}
		}
	}
	
	
	
	/**
	 * Save the state of a table
	 *  @param {object} oSettings dataTables settings object
	 *  @memberof DataTable#oApi
	 */
	function _fnSaveState ( settings )
	{
		if (settings._bLoadingState) {
			return;
		}
	
		/* Store the interesting variables */
		var state = {
			time:    +new Date(),
			start:   settings._iDisplayStart,
			length:  settings._iDisplayLength,
			order:   $.extend( true, [], settings.aaSorting ),
			search:  _fnSearchToCamel( settings.oPreviousSearch ),
			columns: $.map( settings.aoColumns, function ( col, i ) {
				return {
					visible: col.bVisible,
					search: _fnSearchToCamel( settings.aoPreSearchCols[i] )
				};
			} )
		};
	
		settings.oSavedState = state;
		_fnCallbackFire( settings, "aoStateSaveParams", 'stateSaveParams', [settings, state] );
		
		if ( settings.oFeatures.bStateSave && !settings.bDestroying )
		{
			settings.fnStateSaveCallback.call( settings.oInstance, settings, state );
		}	
	}
	
	
	/**
	 * Attempt to load a saved table state
	 *  @param {object} oSettings dataTables settings object
	 *  @param {object} oInit DataTables init object so we can override settings
	 *  @param {function} callback Callback to execute when the state has been loaded
	 *  @memberof DataTable#oApi
	 */
	function _fnLoadState ( settings, oInit, callback )
	{
		if ( ! settings.oFeatures.bStateSave ) {
			callback();
			return;
		}
	
		var loaded = function(state) {
			_fnImplementState(settings, state, callback);
		}
	
		var state = settings.fnStateLoadCallback.call( settings.oInstance, settings, loaded );
	
		if ( state !== undefined ) {
			_fnImplementState( settings, state, callback );
		}
		// otherwise, wait for the loaded callback to be executed
	
		return true;
	}
	
	function _fnImplementState ( settings, s, callback) {
		var i, ien;
		var columns = settings.aoColumns;
		settings._bLoadingState = true;
	
		// When StateRestore was introduced the state could now be implemented at any time
		// Not just initialisation. To do this an api instance is required in some places
		var api = settings._bInitComplete ? new DataTable.Api(settings) : null;
	
		if ( ! s || ! s.time ) {
			settings._bLoadingState = false;
			callback();
			return;
		}
	
		// Allow custom and plug-in manipulation functions to alter the saved data set and
		// cancelling of loading by returning false
		var abStateLoad = _fnCallbackFire( settings, 'aoStateLoadParams', 'stateLoadParams', [settings, s] );
		if ( $.inArray( false, abStateLoad ) !== -1 ) {
			settings._bLoadingState = false;
			callback();
			return;
		}
	
		// Reject old data
		var duration = settings.iStateDuration;
		if ( duration > 0 && s.time < +new Date() - (duration*1000) ) {
			settings._bLoadingState = false;
			callback();
			return;
		}
	
		// Number of columns have changed - all bets are off, no restore of settings
		if ( s.columns && columns.length !== s.columns.length ) {
			settings._bLoadingState = false;
			callback();
			return;
		}
	
		// Store the saved state so it might be accessed at any time
		settings.oLoadedState = $.extend( true, {}, s );
	
		// Page Length
		if ( s.length !== undefined ) {
			// If already initialised just set the value directly so that the select element is also updated
			if (api) {
				api.page.len(s.length)
			}
			else {
				settings._iDisplayLength   = s.length;
			}
		}
	
		// Restore key features - todo - for 1.11 this needs to be done by
		// subscribed events
		if ( s.start !== undefined ) {
			if(api === null) {
				settings._iDisplayStart    = s.start;
				settings.iInitDisplayStart = s.start;
			}
			else {
				_fnPageChange(settings, s.start/settings._iDisplayLength);
			}
		}
	
		// Order
		if ( s.order !== undefined ) {
			settings.aaSorting = [];
			$.each( s.order, function ( i, col ) {
				settings.aaSorting.push( col[0] >= columns.length ?
					[ 0, col[1] ] :
					col
				);
			} );
		}
	
		// Search
		if ( s.search !== undefined ) {
			$.extend( settings.oPreviousSearch, _fnSearchToHung( s.search ) );
		}
	
		// Columns
		if ( s.columns ) {
			for ( i=0, ien=s.columns.length ; i<ien ; i++ ) {
				var col = s.columns[i];
	
				// Visibility
				if ( col.visible !== undefined ) {
					// If the api is defined, the table has been initialised so we need to use it rather than internal settings
					if (api) {
						// Don't redraw the columns on every iteration of this loop, we will do this at the end instead
						api.column(i).visible(col.visible, false);
					}
					else {
						columns[i].bVisible = col.visible;
					}
				}
	
				// Search
				if ( col.search !== undefined ) {
					$.extend( settings.aoPreSearchCols[i], _fnSearchToHung( col.search ) );
				}
			}
			
			// If the api is defined then we need to adjust the columns once the visibility has been changed
			if (api) {
				api.columns.adjust();
			}
		}
	
		settings._bLoadingState = false;
		_fnCallbackFire( settings, 'aoStateLoaded', 'stateLoaded', [settings, s] );
		callback();
	};
	
	
	/**
	 * Return the settings object for a particular table
	 *  @param {node} table table we are using as a dataTable
	 *  @returns {object} Settings object - or null if not found
	 *  @memberof DataTable#oApi
	 */
	function _fnSettingsFromNode ( table )
	{
		var settings = DataTable.settings;
		var idx = $.inArray( table, _pluck( settings, 'nTable' ) );
	
		return idx !== -1 ?
			settings[ idx ] :
			null;
	}
	
	
	/**
	 * Log an error message
	 *  @param {object} settings dataTables settings object
	 *  @param {int} level log error messages, or display them to the user
	 *  @param {string} msg error message
	 *  @param {int} tn Technical note id to get more information about the error.
	 *  @memberof DataTable#oApi
	 */
	function _fnLog( settings, level, msg, tn )
	{
		msg = 'DataTables warning: '+
			(settings ? 'table id='+settings.sTableId+' - ' : '')+msg;
	
		if ( tn ) {
			msg += '. For more information about this error, please see '+
			'http://datatables.net/tn/'+tn;
		}
	
		if ( ! level  ) {
			// Backwards compatibility pre 1.10
			var ext = DataTable.ext;
			var type = ext.sErrMode || ext.errMode;
	
			if ( settings ) {
				_fnCallbackFire( settings, null, 'error', [ settings, tn, msg ] );
			}
	
			if ( type == 'alert' ) {
				alert( msg );
			}
			else if ( type == 'throw' ) {
				throw new Error(msg);
			}
			else if ( typeof type == 'function' ) {
				type( settings, tn, msg );
			}
		}
		else if ( window.console && console.log ) {
			console.log( msg );
		}
	}
	
	
	/**
	 * See if a property is defined on one object, if so assign it to the other object
	 *  @param {object} ret target object
	 *  @param {object} src source object
	 *  @param {string} name property
	 *  @param {string} [mappedName] name to map too - optional, name used if not given
	 *  @memberof DataTable#oApi
	 */
	function _fnMap( ret, src, name, mappedName )
	{
		if ( Array.isArray( name ) ) {
			$.each( name, function (i, val) {
				if ( Array.isArray( val ) ) {
					_fnMap( ret, src, val[0], val[1] );
				}
				else {
					_fnMap( ret, src, val );
				}
			} );
	
			return;
		}
	
		if ( mappedName === undefined ) {
			mappedName = name;
		}
	
		if ( src[name] !== undefined ) {
			ret[mappedName] = src[name];
		}
	}
	
	
	/**
	 * Extend objects - very similar to jQuery.extend, but deep copy objects, and
	 * shallow copy arrays. The reason we need to do this, is that we don't want to
	 * deep copy array init values (such as aaSorting) since the dev wouldn't be
	 * able to override them, but we do want to deep copy arrays.
	 *  @param {object} out Object to extend
	 *  @param {object} extender Object from which the properties will be applied to
	 *      out
	 *  @param {boolean} breakRefs If true, then arrays will be sliced to take an
	 *      independent copy with the exception of the `data` or `aaData` parameters
	 *      if they are present. This is so you can pass in a collection to
	 *      DataTables and have that used as your data source without breaking the
	 *      references
	 *  @returns {object} out Reference, just for convenience - out === the return.
	 *  @memberof DataTable#oApi
	 *  @todo This doesn't take account of arrays inside the deep copied objects.
	 */
	function _fnExtend( out, extender, breakRefs )
	{
		var val;
	
		for ( var prop in extender ) {
			if ( extender.hasOwnProperty(prop) ) {
				val = extender[prop];
	
				if ( $.isPlainObject( val ) ) {
					if ( ! $.isPlainObject( out[prop] ) ) {
						out[prop] = {};
					}
					$.extend( true, out[prop], val );
				}
				else if ( breakRefs && prop !== 'data' && prop !== 'aaData' && Array.isArray(val) ) {
					out[prop] = val.slice();
				}
				else {
					out[prop] = val;
				}
			}
		}
	
		return out;
	}
	
	
	/**
	 * Bind an event handers to allow a click or return key to activate the callback.
	 * This is good for accessibility since a return on the keyboard will have the
	 * same effect as a click, if the element has focus.
	 *  @param {element} n Element to bind the action to
	 *  @param {object} oData Data object to pass to the triggered function
	 *  @param {function} fn Callback function for when the event is triggered
	 *  @memberof DataTable#oApi
	 */
	function _fnBindAction( n, oData, fn )
	{
		$(n)
			.on( 'click.DT', oData, function (e) {
					$(n).trigger('blur'); // Remove focus outline for mouse users
					fn(e);
				} )
			.on( 'keypress.DT', oData, function (e){
					if ( e.which === 13 ) {
						e.preventDefault();
						fn(e);
					}
				} )
			.on( 'selectstart.DT', function () {
					/* Take the brutal approach to cancelling text selection */
					return false;
				} );
	}
	
	
	/**
	 * Register a callback function. Easily allows a callback function to be added to
	 * an array store of callback functions that can then all be called together.
	 *  @param {object} oSettings dataTables settings object
	 *  @param {string} sStore Name of the array storage for the callbacks in oSettings
	 *  @param {function} fn Function to be called back
	 *  @param {string} sName Identifying name for the callback (i.e. a label)
	 *  @memberof DataTable#oApi
	 */
	function _fnCallbackReg( oSettings, sStore, fn, sName )
	{
		if ( fn )
		{
			oSettings[sStore].push( {
				"fn": fn,
				"sName": sName
			} );
		}
	}
	
	
	/**
	 * Fire callback functions and trigger events. Note that the loop over the
	 * callback array store is done backwards! Further note that you do not want to
	 * fire off triggers in time sensitive applications (for example cell creation)
	 * as its slow.
	 *  @param {object} settings dataTables settings object
	 *  @param {string} callbackArr Name of the array storage for the callbacks in
	 *      oSettings
	 *  @param {string} eventName Name of the jQuery custom event to trigger. If
	 *      null no trigger is fired
	 *  @param {array} args Array of arguments to pass to the callback function /
	 *      trigger
	 *  @memberof DataTable#oApi
	 */
	function _fnCallbackFire( settings, callbackArr, eventName, args )
	{
		var ret = [];
	
		if ( callbackArr ) {
			ret = $.map( settings[callbackArr].slice().reverse(), function (val, i) {
				return val.fn.apply( settings.oInstance, args );
			} );
		}
	
		if ( eventName !== null ) {
			var e = $.Event( eventName+'.dt' );
			var table = $(settings.nTable);
	
			table.trigger( e, args );
	
			// If not yet attached to the document, trigger the event
			// on the body directly to sort of simulate the bubble
			if (table.parents('body').length === 0) {
				$('body').trigger( e, args );
			}
	
			ret.push( e.result );
		}
	
		return ret;
	}
	
	
	function _fnLengthOverflow ( settings )
	{
		var
			start = settings._iDisplayStart,
			end = settings.fnDisplayEnd(),
			len = settings._iDisplayLength;
	
		/* If we have space to show extra rows (backing up from the end point - then do so */
		if ( start >= end )
		{
			start = end - len;
		}
	
		// Keep the start record on the current page
		start -= (start % len);
	
		if ( len === -1 || start < 0 )
		{
			start = 0;
		}
	
		settings._iDisplayStart = start;
	}
	
	
	function _fnRenderer( settings, type )
	{
		var renderer = settings.renderer;
		var host = DataTable.ext.renderer[type];
	
		if ( $.isPlainObject( renderer ) && renderer[type] ) {
			// Specific renderer for this type. If available use it, otherwise use
			// the default.
			return host[renderer[type]] || host._;
		}
		else if ( typeof renderer === 'string' ) {
			// Common renderer - if there is one available for this type use it,
			// otherwise use the default
			return host[renderer] || host._;
		}
	
		// Use the default
		return host._;
	}
	
	
	/**
	 * Detect the data source being used for the table. Used to simplify the code
	 * a little (ajax) and to make it compress a little smaller.
	 *
	 *  @param {object} settings dataTables settings object
	 *  @returns {string} Data source
	 *  @memberof DataTable#oApi
	 */
	function _fnDataSource ( settings )
	{
		if ( settings.oFeatures.bServerSide ) {
			return 'ssp';
		}
		else if ( settings.ajax || settings.sAjaxSource ) {
			return 'ajax';
		}
		return 'dom';
	}
	
	
	
	
	/**
	 * Computed structure of the DataTables API, defined by the options passed to
	 * `DataTable.Api.register()` when building the API.
	 *
	 * The structure is built in order to speed creation and extension of the Api
	 * objects since the extensions are effectively pre-parsed.
	 *
	 * The array is an array of objects with the following structure, where this
	 * base array represents the Api prototype base:
	 *
	 *     [
	 *       {
	 *         name:      'data'                -- string   - Property name
	 *         val:       function () {},       -- function - Api method (or undefined if just an object
	 *         methodExt: [ ... ],              -- array    - Array of Api object definitions to extend the method result
	 *         propExt:   [ ... ]               -- array    - Array of Api object definitions to extend the property
	 *       },
	 *       {
	 *         name:     'row'
	 *         val:       {},
	 *         methodExt: [ ... ],
	 *         propExt:   [
	 *           {
	 *             name:      'data'
	 *             val:       function () {},
	 *             methodExt: [ ... ],
	 *             propExt:   [ ... ]
	 *           },
	 *           ...
	 *         ]
	 *       }
	 *     ]
	 *
	 * @type {Array}
	 * @ignore
	 */
	var __apiStruct = [];
	
	
	/**
	 * `Array.prototype` reference.
	 *
	 * @type object
	 * @ignore
	 */
	var __arrayProto = Array.prototype;
	
	
	/**
	 * Abstraction for `context` parameter of the `Api` constructor to allow it to
	 * take several different forms for ease of use.
	 *
	 * Each of the input parameter types will be converted to a DataTables settings
	 * object where possible.
	 *
	 * @param  {string|node|jQuery|object} mixed DataTable identifier. Can be one
	 *   of:
	 *
	 *   * `string` - jQuery selector. Any DataTables' matching the given selector
	 *     with be found and used.
	 *   * `node` - `TABLE` node which has already been formed into a DataTable.
	 *   * `jQuery` - A jQuery object of `TABLE` nodes.
	 *   * `object` - DataTables settings object
	 *   * `DataTables.Api` - API instance
	 * @return {array|null} Matching DataTables settings objects. `null` or
	 *   `undefined` is returned if no matching DataTable is found.
	 * @ignore
	 */
	var _toSettings = function ( mixed )
	{
		var idx, jq;
		var settings = DataTable.settings;
		var tables = $.map( settings, function (el, i) {
			return el.nTable;
		} );
	
		if ( ! mixed ) {
			return [];
		}
		else if ( mixed.nTable && mixed.oApi ) {
			// DataTables settings object
			return [ mixed ];
		}
		else if ( mixed.nodeName && mixed.nodeName.toLowerCase() === 'table' ) {
			// Table node
			idx = $.inArray( mixed, tables );
			return idx !== -1 ? [ settings[idx] ] : null;
		}
		else if ( mixed && typeof mixed.settings === 'function' ) {
			return mixed.settings().toArray();
		}
		else if ( typeof mixed === 'string' ) {
			// jQuery selector
			jq = $(mixed);
		}
		else if ( mixed instanceof $ ) {
			// jQuery object (also DataTables instance)
			jq = mixed;
		}
	
		if ( jq ) {
			return jq.map( function(i) {
				idx = $.inArray( this, tables );
				return idx !== -1 ? settings[idx] : null;
			} ).toArray();
		}
	};
	
	
	/**
	 * DataTables API class - used to control and interface with  one or more
	 * DataTables enhanced tables.
	 *
	 * The API class is heavily based on jQuery, presenting a chainable interface
	 * that you can use to interact with tables. Each instance of the API class has
	 * a "context" - i.e. the tables that it will operate on. This could be a single
	 * table, all tables on a page or a sub-set thereof.
	 *
	 * Additionally the API is designed to allow you to easily work with the data in
	 * the tables, retrieving and manipulating it as required. This is done by
	 * presenting the API class as an array like interface. The contents of the
	 * array depend upon the actions requested by each method (for example
	 * `rows().nodes()` will return an array of nodes, while `rows().data()` will
	 * return an array of objects or arrays depending upon your table's
	 * configuration). The API object has a number of array like methods (`push`,
	 * `pop`, `reverse` etc) as well as additional helper methods (`each`, `pluck`,
	 * `unique` etc) to assist your working with the data held in a table.
	 *
	 * Most methods (those which return an Api instance) are chainable, which means
	 * the return from a method call also has all of the methods available that the
	 * top level object had. For example, these two calls are equivalent:
	 *
	 *     // Not chained
	 *     api.row.add( {...} );
	 *     api.draw();
	 *
	 *     // Chained
	 *     api.row.add( {...} ).draw();
	 *
	 * @class DataTable.Api
	 * @param {array|object|string|jQuery} context DataTable identifier. This is
	 *   used to define which DataTables enhanced tables this API will operate on.
	 *   Can be one of:
	 *
	 *   * `string` - jQuery selector. Any DataTables' matching the given selector
	 *     with be found and used.
	 *   * `node` - `TABLE` node which has already been formed into a DataTable.
	 *   * `jQuery` - A jQuery object of `TABLE` nodes.
	 *   * `object` - DataTables settings object
	 * @param {array} [data] Data to initialise the Api instance with.
	 *
	 * @example
	 *   // Direct initialisation during DataTables construction
	 *   var api = $('#example').DataTable();
	 *
	 * @example
	 *   // Initialisation using a DataTables jQuery object
	 *   var api = $('#example').dataTable().api();
	 *
	 * @example
	 *   // Initialisation as a constructor
	 *   var api = new $.fn.DataTable.Api( 'table.dataTable' );
	 */
	_Api = function ( context, data )
	{
		if ( ! (this instanceof _Api) ) {
			return new _Api( context, data );
		}
	
		var settings = [];
		var ctxSettings = function ( o ) {
			var a = _toSettings( o );
			if ( a ) {
				settings.push.apply( settings, a );
			}
		};
	
		if ( Array.isArray( context ) ) {
			for ( var i=0, ien=context.length ; i<ien ; i++ ) {
				ctxSettings( context[i] );
			}
		}
		else {
			ctxSettings( context );
		}
	
		// Remove duplicates
		this.context = _unique( settings );
	
		// Initial data
		if ( data ) {
			$.merge( this, data );
		}
	
		// selector
		this.selector = {
			rows: null,
			cols: null,
			opts: null
		};
	
		_Api.extend( this, this, __apiStruct );
	};
	
	DataTable.Api = _Api;
	
	// Don't destroy the existing prototype, just extend it. Required for jQuery 2's
	// isPlainObject.
	$.extend( _Api.prototype, {
		any: function ()
		{
			return this.count() !== 0;
		},
	
	
		concat:  __arrayProto.concat,
	
	
		context: [], // array of table settings objects
	
	
		count: function ()
		{
			return this.flatten().length;
		},
	
	
		each: function ( fn )
		{
			for ( var i=0, ien=this.length ; i<ien; i++ ) {
				fn.call( this, this[i], i, this );
			}
	
			return this;
		},
	
	
		eq: function ( idx )
		{
			var ctx = this.context;
	
			return ctx.length > idx ?
				new _Api( ctx[idx], this[idx] ) :
				null;
		},
	
	
		filter: function ( fn )
		{
			var a = [];
	
			if ( __arrayProto.filter ) {
				a = __arrayProto.filter.call( this, fn, this );
			}
			else {
				// Compatibility for browsers without EMCA-252-5 (JS 1.6)
				for ( var i=0, ien=this.length ; i<ien ; i++ ) {
					if ( fn.call( this, this[i], i, this ) ) {
						a.push( this[i] );
					}
				}
			}
	
			return new _Api( this.context, a );
		},
	
	
		flatten: function ()
		{
			var a = [];
			return new _Api( this.context, a.concat.apply( a, this.toArray() ) );
		},
	
	
		join:    __arrayProto.join,
	
	
		indexOf: __arrayProto.indexOf || function (obj, start)
		{
			for ( var i=(start || 0), ien=this.length ; i<ien ; i++ ) {
				if ( this[i] === obj ) {
					return i;
				}
			}
			return -1;
		},
	
		iterator: function ( flatten, type, fn, alwaysNew ) {
			var
				a = [], ret,
				i, ien, j, jen,
				context = this.context,
				rows, items, item,
				selector = this.selector;
	
			// Argument shifting
			if ( typeof flatten === 'string' ) {
				alwaysNew = fn;
				fn = type;
				type = flatten;
				flatten = false;
			}
	
			for ( i=0, ien=context.length ; i<ien ; i++ ) {
				var apiInst = new _Api( context[i] );
	
				if ( type === 'table' ) {
					ret = fn.call( apiInst, context[i], i );
	
					if ( ret !== undefined ) {
						a.push( ret );
					}
				}
				else if ( type === 'columns' || type === 'rows' ) {
					// this has same length as context - one entry for each table
					ret = fn.call( apiInst, context[i], this[i], i );
	
					if ( ret !== undefined ) {
						a.push( ret );
					}
				}
				else if ( type === 'column' || type === 'column-rows' || type === 'row' || type === 'cell' ) {
					// columns and rows share the same structure.
					// 'this' is an array of column indexes for each context
					items = this[i];
	
					if ( type === 'column-rows' ) {
						rows = _selector_row_indexes( context[i], selector.opts );
					}
	
					for ( j=0, jen=items.length ; j<jen ; j++ ) {
						item = items[j];
	
						if ( type === 'cell' ) {
							ret = fn.call( apiInst, context[i], item.row, item.column, i, j );
						}
						else {
							ret = fn.call( apiInst, context[i], item, i, j, rows );
						}
	
						if ( ret !== undefined ) {
							a.push( ret );
						}
					}
				}
			}
	
			if ( a.length || alwaysNew ) {
				var api = new _Api( context, flatten ? a.concat.apply( [], a ) : a );
				var apiSelector = api.selector;
				apiSelector.rows = selector.rows;
				apiSelector.cols = selector.cols;
				apiSelector.opts = selector.opts;
				return api;
			}
			return this;
		},
	
	
		lastIndexOf: __arrayProto.lastIndexOf || function (obj, start)
		{
			// Bit cheeky...
			return this.indexOf.apply( this.toArray.reverse(), arguments );
		},
	
	
		length:  0,
	
	
		map: function ( fn )
		{
			var a = [];
	
			if ( __arrayProto.map ) {
				a = __arrayProto.map.call( this, fn, this );
			}
			else {
				// Compatibility for browsers without EMCA-252-5 (JS 1.6)
				for ( var i=0, ien=this.length ; i<ien ; i++ ) {
					a.push( fn.call( this, this[i], i ) );
				}
			}
	
			return new _Api( this.context, a );
		},
	
	
		pluck: function ( prop )
		{
			var fn = DataTable.util.get(prop);
	
			return this.map( function ( el ) {
				return fn(el);
			} );
		},
	
		pop:     __arrayProto.pop,
	
	
		push:    __arrayProto.push,
	
	
		// Does not return an API instance
		reduce: __arrayProto.reduce || function ( fn, init )
		{
			return _fnReduce( this, fn, init, 0, this.length, 1 );
		},
	
	
		reduceRight: __arrayProto.reduceRight || function ( fn, init )
		{
			return _fnReduce( this, fn, init, this.length-1, -1, -1 );
		},
	
	
		reverse: __arrayProto.reverse,
	
	
		// Object with rows, columns and opts
		selector: null,
	
	
		shift:   __arrayProto.shift,
	
	
		slice: function () {
			return new _Api( this.context, this );
		},
	
	
		sort:    __arrayProto.sort, // ? name - order?
	
	
		splice:  __arrayProto.splice,
	
	
		toArray: function ()
		{
			return __arrayProto.slice.call( this );
		},
	
	
		to$: function ()
		{
			return $( this );
		},
	
	
		toJQuery: function ()
		{
			return $( this );
		},
	
	
		unique: function ()
		{
			return new _Api( this.context, _unique(this) );
		},
	
	
		unshift: __arrayProto.unshift
	} );
	
	
	_Api.extend = function ( scope, obj, ext )
	{
		// Only extend API instances and static properties of the API
		if ( ! ext.length || ! obj || ( ! (obj instanceof _Api) && ! obj.__dt_wrapper ) ) {
			return;
		}
	
		var
			i, ien,
			struct,
			methodScoping = function ( scope, fn, struc ) {
				return function () {
					var ret = fn.apply( scope, arguments );
	
					// Method extension
					_Api.extend( ret, ret, struc.methodExt );
					return ret;
				};
			};
	
		for ( i=0, ien=ext.length ; i<ien ; i++ ) {
			struct = ext[i];
	
			// Value
			obj[ struct.name ] = struct.type === 'function' ?
				methodScoping( scope, struct.val, struct ) :
				struct.type === 'object' ?
					{} :
					struct.val;
	
			obj[ struct.name ].__dt_wrapper = true;
	
			// Property extension
			_Api.extend( scope, obj[ struct.name ], struct.propExt );
		}
	};
	
	
	// @todo - Is there need for an augment function?
	// _Api.augment = function ( inst, name )
	// {
	// 	// Find src object in the structure from the name
	// 	var parts = name.split('.');
	
	// 	_Api.extend( inst, obj );
	// };
	
	
	//     [
	//       {
	//         name:      'data'                -- string   - Property name
	//         val:       function () {},       -- function - Api method (or undefined if just an object
	//         methodExt: [ ... ],              -- array    - Array of Api object definitions to extend the method result
	//         propExt:   [ ... ]               -- array    - Array of Api object definitions to extend the property
	//       },
	//       {
	//         name:     'row'
	//         val:       {},
	//         methodExt: [ ... ],
	//         propExt:   [
	//           {
	//             name:      'data'
	//             val:       function () {},
	//             methodExt: [ ... ],
	//             propExt:   [ ... ]
	//           },
	//           ...
	//         ]
	//       }
	//     ]
	
	_Api.register = _api_register = function ( name, val )
	{
		if ( Array.isArray( name ) ) {
			for ( var j=0, jen=name.length ; j<jen ; j++ ) {
				_Api.register( name[j], val );
			}
			return;
		}
	
		var
			i, ien,
			heir = name.split('.'),
			struct = __apiStruct,
			key, method;
	
		var find = function ( src, name ) {
			for ( var i=0, ien=src.length ; i<ien ; i++ ) {
				if ( src[i].name === name ) {
					return src[i];
				}
			}
			return null;
		};
	
		for ( i=0, ien=heir.length ; i<ien ; i++ ) {
			method = heir[i].indexOf('()') !== -1;
			key = method ?
				heir[i].replace('()', '') :
				heir[i];
	
			var src = find( struct, key );
			if ( ! src ) {
				src = {
					name:      key,
					val:       {},
					methodExt: [],
					propExt:   [],
					type:      'object'
				};
				struct.push( src );
			}
	
			if ( i === ien-1 ) {
				src.val = val;
				src.type = typeof val === 'function' ?
					'function' :
					$.isPlainObject( val ) ?
						'object' :
						'other';
			}
			else {
				struct = method ?
					src.methodExt :
					src.propExt;
			}
		}
	};
	
	_Api.registerPlural = _api_registerPlural = function ( pluralName, singularName, val ) {
		_Api.register( pluralName, val );
	
		_Api.register( singularName, function () {
			var ret = val.apply( this, arguments );
	
			if ( ret === this ) {
				// Returned item is the API instance that was passed in, return it
				return this;
			}
			else if ( ret instanceof _Api ) {
				// New API instance returned, want the value from the first item
				// in the returned array for the singular result.
				return ret.length ?
					Array.isArray( ret[0] ) ?
						new _Api( ret.context, ret[0] ) : // Array results are 'enhanced'
						ret[0] :
					undefined;
			}
	
			// Non-API return - just fire it back
			return ret;
		} );
	};
	
	
	/**
	 * Selector for HTML tables. Apply the given selector to the give array of
	 * DataTables settings objects.
	 *
	 * @param {string|integer} [selector] jQuery selector string or integer
	 * @param  {array} Array of DataTables settings objects to be filtered
	 * @return {array}
	 * @ignore
	 */
	var __table_selector = function ( selector, a )
	{
		if ( Array.isArray(selector) ) {
			return $.map( selector, function (item) {
				return __table_selector(item, a);
			} );
		}
	
		// Integer is used to pick out a table by index
		if ( typeof selector === 'number' ) {
			return [ a[ selector ] ];
		}
	
		// Perform a jQuery selector on the table nodes
		var nodes = $.map( a, function (el, i) {
			return el.nTable;
		} );
	
		return $(nodes)
			.filter( selector )
			.map( function (i) {
				// Need to translate back from the table node to the settings
				var idx = $.inArray( this, nodes );
				return a[ idx ];
			} )
			.toArray();
	};
	
	
	
	/**
	 * Context selector for the API's context (i.e. the tables the API instance
	 * refers to.
	 *
	 * @name    DataTable.Api#tables
	 * @param {string|integer} [selector] Selector to pick which tables the iterator
	 *   should operate on. If not given, all tables in the current context are
	 *   used. This can be given as a jQuery selector (for example `':gt(0)'`) to
	 *   select multiple tables or as an integer to select a single table.
	 * @returns {DataTable.Api} Returns a new API instance if a selector is given.
	 */
	_api_register( 'tables()', function ( selector ) {
		// A new instance is created if there was a selector specified
		return selector !== undefined && selector !== null ?
			new _Api( __table_selector( selector, this.context ) ) :
			this;
	} );
	
	
	_api_register( 'table()', function ( selector ) {
		var tables = this.tables( selector );
		var ctx = tables.context;
	
		// Truncate to the first matched table
		return ctx.length ?
			new _Api( ctx[0] ) :
			tables;
	} );
	
	
	_api_registerPlural( 'tables().nodes()', 'table().node()' , function () {
		return this.iterator( 'table', function ( ctx ) {
			return ctx.nTable;
		}, 1 );
	} );
	
	
	_api_registerPlural( 'tables().body()', 'table().body()' , function () {
		return this.iterator( 'table', function ( ctx ) {
			return ctx.nTBody;
		}, 1 );
	} );
	
	
	_api_registerPlural( 'tables().header()', 'table().header()' , function () {
		return this.iterator( 'table', function ( ctx ) {
			return ctx.nTHead;
		}, 1 );
	} );
	
	
	_api_registerPlural( 'tables().footer()', 'table().footer()' , function () {
		return this.iterator( 'table', function ( ctx ) {
			return ctx.nTFoot;
		}, 1 );
	} );
	
	
	_api_registerPlural( 'tables().containers()', 'table().container()' , function () {
		return this.iterator( 'table', function ( ctx ) {
			return ctx.nTableWrapper;
		}, 1 );
	} );
	
	
	
	/**
	 * Redraw the tables in the current context.
	 */
	_api_register( 'draw()', function ( paging ) {
		return this.iterator( 'table', function ( settings ) {
			if ( paging === 'page' ) {
				_fnDraw( settings );
			}
			else {
				if ( typeof paging === 'string' ) {
					paging = paging === 'full-hold' ?
						false :
						true;
				}
	
				_fnReDraw( settings, paging===false );
			}
		} );
	} );
	
	
	
	/**
	 * Get the current page index.
	 *
	 * @return {integer} Current page index (zero based)
	 *//**
	 * Set the current page.
	 *
	 * Note that if you attempt to show a page which does not exist, DataTables will
	 * not throw an error, but rather reset the paging.
	 *
	 * @param {integer|string} action The paging action to take. This can be one of:
	 *  * `integer` - The page index to jump to
	 *  * `string` - An action to take:
	 *    * `first` - Jump to first page.
	 *    * `next` - Jump to the next page
	 *    * `previous` - Jump to previous page
	 *    * `last` - Jump to the last page.
	 * @returns {DataTables.Api} this
	 */
	_api_register( 'page()', function ( action ) {
		if ( action === undefined ) {
			return this.page.info().page; // not an expensive call
		}
	
		// else, have an action to take on all tables
		return this.iterator( 'table', function ( settings ) {
			_fnPageChange( settings, action );
		} );
	} );
	
	
	/**
	 * Paging information for the first table in the current context.
	 *
	 * If you require paging information for another table, use the `table()` method
	 * with a suitable selector.
	 *
	 * @return {object} Object with the following properties set:
	 *  * `page` - Current page index (zero based - i.e. the first page is `0`)
	 *  * `pages` - Total number of pages
	 *  * `start` - Display index for the first record shown on the current page
	 *  * `end` - Display index for the last record shown on the current page
	 *  * `length` - Display length (number of records). Note that generally `start
	 *    + length = end`, but this is not always true, for example if there are
	 *    only 2 records to show on the final page, with a length of 10.
	 *  * `recordsTotal` - Full data set length
	 *  * `recordsDisplay` - Data set length once the current filtering criterion
	 *    are applied.
	 */
	_api_register( 'page.info()', function ( action ) {
		if ( this.context.length === 0 ) {
			return undefined;
		}
	
		var
			settings   = this.context[0],
			start      = settings._iDisplayStart,
			len        = settings.oFeatures.bPaginate ? settings._iDisplayLength : -1,
			visRecords = settings.fnRecordsDisplay(),
			all        = len === -1;
	
		return {
			"page":           all ? 0 : Math.floor( start / len ),
			"pages":          all ? 1 : Math.ceil( visRecords / len ),
			"start":          start,
			"end":            settings.fnDisplayEnd(),
			"length":         len,
			"recordsTotal":   settings.fnRecordsTotal(),
			"recordsDisplay": visRecords,
			"serverSide":     _fnDataSource( settings ) === 'ssp'
		};
	} );
	
	
	/**
	 * Get the current page length.
	 *
	 * @return {integer} Current page length. Note `-1` indicates that all records
	 *   are to be shown.
	 *//**
	 * Set the current page length.
	 *
	 * @param {integer} Page length to set. Use `-1` to show all records.
	 * @returns {DataTables.Api} this
	 */
	_api_register( 'page.len()', function ( len ) {
		// Note that we can't call this function 'length()' because `length`
		// is a Javascript property of functions which defines how many arguments
		// the function expects.
		if ( len === undefined ) {
			return this.context.length !== 0 ?
				this.context[0]._iDisplayLength :
				undefined;
		}
	
		// else, set the page length
		return this.iterator( 'table', function ( settings ) {
			_fnLengthChange( settings, len );
		} );
	} );
	
	
	
	var __reload = function ( settings, holdPosition, callback ) {
		// Use the draw event to trigger a callback
		if ( callback ) {
			var api = new _Api( settings );
	
			api.one( 'draw', function () {
				callback( api.ajax.json() );
			} );
		}
	
		if ( _fnDataSource( settings ) == 'ssp' ) {
			_fnReDraw( settings, holdPosition );
		}
		else {
			_fnProcessingDisplay( settings, true );
	
			// Cancel an existing request
			var xhr = settings.jqXHR;
			if ( xhr && xhr.readyState !== 4 ) {
				xhr.abort();
			}
	
			// Trigger xhr
			_fnBuildAjax( settings, [], function( json ) {
				_fnClearTable( settings );
	
				var data = _fnAjaxDataSrc( settings, json );
				for ( var i=0, ien=data.length ; i<ien ; i++ ) {
					_fnAddData( settings, data[i] );
				}
	
				_fnReDraw( settings, holdPosition );
				_fnProcessingDisplay( settings, false );
			} );
		}
	};
	
	
	/**
	 * Get the JSON response from the last Ajax request that DataTables made to the
	 * server. Note that this returns the JSON from the first table in the current
	 * context.
	 *
	 * @return {object} JSON received from the server.
	 */
	_api_register( 'ajax.json()', function () {
		var ctx = this.context;
	
		if ( ctx.length > 0 ) {
			return ctx[0].json;
		}
	
		// else return undefined;
	} );
	
	
	/**
	 * Get the data submitted in the last Ajax request
	 */
	_api_register( 'ajax.params()', function () {
		var ctx = this.context;
	
		if ( ctx.length > 0 ) {
			return ctx[0].oAjaxData;
		}
	
		// else return undefined;
	} );
	
	
	/**
	 * Reload tables from the Ajax data source. Note that this function will
	 * automatically re-draw the table when the remote data has been loaded.
	 *
	 * @param {boolean} [reset=true] Reset (default) or hold the current paging
	 *   position. A full re-sort and re-filter is performed when this method is
	 *   called, which is why the pagination reset is the default action.
	 * @returns {DataTables.Api} this
	 */
	_api_register( 'ajax.reload()', function ( callback, resetPaging ) {
		return this.iterator( 'table', function (settings) {
			__reload( settings, resetPaging===false, callback );
		} );
	} );
	
	
	/**
	 * Get the current Ajax URL. Note that this returns the URL from the first
	 * table in the current context.
	 *
	 * @return {string} Current Ajax source URL
	 *//**
	 * Set the Ajax URL. Note that this will set the URL for all tables in the
	 * current context.
	 *
	 * @param {string} url URL to set.
	 * @returns {DataTables.Api} this
	 */
	_api_register( 'ajax.url()', function ( url ) {
		var ctx = this.context;
	
		if ( url === undefined ) {
			// get
			if ( ctx.length === 0 ) {
				return undefined;
			}
			ctx = ctx[0];
	
			return ctx.ajax ?
				$.isPlainObject( ctx.ajax ) ?
					ctx.ajax.url :
					ctx.ajax :
				ctx.sAjaxSource;
		}
	
		// set
		return this.iterator( 'table', function ( settings ) {
			if ( $.isPlainObject( settings.ajax ) ) {
				settings.ajax.url = url;
			}
			else {
				settings.ajax = url;
			}
			// No need to consider sAjaxSource here since DataTables gives priority
			// to `ajax` over `sAjaxSource`. So setting `ajax` here, renders any
			// value of `sAjaxSource` redundant.
		} );
	} );
	
	
	/**
	 * Load data from the newly set Ajax URL. Note that this method is only
	 * available when `ajax.url()` is used to set a URL. Additionally, this method
	 * has the same effect as calling `ajax.reload()` but is provided for
	 * convenience when setting a new URL. Like `ajax.reload()` it will
	 * automatically redraw the table once the remote data has been loaded.
	 *
	 * @returns {DataTables.Api} this
	 */
	_api_register( 'ajax.url().load()', function ( callback, resetPaging ) {
		// Same as a reload, but makes sense to present it for easy access after a
		// url change
		return this.iterator( 'table', function ( ctx ) {
			__reload( ctx, resetPaging===false, callback );
		} );
	} );
	
	
	
	
	var _selector_run = function ( type, selector, selectFn, settings, opts )
	{
		var
			out = [], res,
			a, i, ien, j, jen,
			selectorType = typeof selector;
	
		// Can't just check for isArray here, as an API or jQuery instance might be
		// given with their array like look
		if ( ! selector || selectorType === 'string' || selectorType === 'function' || selector.length === undefined ) {
			selector = [ selector ];
		}
	
		for ( i=0, ien=selector.length ; i<ien ; i++ ) {
			// Only split on simple strings - complex expressions will be jQuery selectors
			a = selector[i] && selector[i].split && ! selector[i].match(/[\[\(:]/) ?
				selector[i].split(',') :
				[ selector[i] ];
	
			for ( j=0, jen=a.length ; j<jen ; j++ ) {
				res = selectFn( typeof a[j] === 'string' ? (a[j]).trim() : a[j] );
	
				if ( res && res.length ) {
					out = out.concat( res );
				}
			}
		}
	
		// selector extensions
		var ext = _ext.selector[ type ];
		if ( ext.length ) {
			for ( i=0, ien=ext.length ; i<ien ; i++ ) {
				out = ext[i]( settings, opts, out );
			}
		}
	
		return _unique( out );
	};
	
	
	var _selector_opts = function ( opts )
	{
		if ( ! opts ) {
			opts = {};
		}
	
		// Backwards compatibility for 1.9- which used the terminology filter rather
		// than search
		if ( opts.filter && opts.search === undefined ) {
			opts.search = opts.filter;
		}
	
		return $.extend( {
			search: 'none',
			order: 'current',
			page: 'all'
		}, opts );
	};
	
	
	var _selector_first = function ( inst )
	{
		// Reduce the API instance to the first item found
		for ( var i=0, ien=inst.length ; i<ien ; i++ ) {
			if ( inst[i].length > 0 ) {
				// Assign the first element to the first item in the instance
				// and truncate the instance and context
				inst[0] = inst[i];
				inst[0].length = 1;
				inst.length = 1;
				inst.context = [ inst.context[i] ];
	
				return inst;
			}
		}
	
		// Not found - return an empty instance
		inst.length = 0;
		return inst;
	};
	
	
	var _selector_row_indexes = function ( settings, opts )
	{
		var
			i, ien, tmp, a=[],
			displayFiltered = settings.aiDisplay,
			displayMaster = settings.aiDisplayMaster;
	
		var
			search = opts.search,  // none, applied, removed
			order  = opts.order,   // applied, current, index (original - compatibility with 1.9)
			page   = opts.page;    // all, current
	
		if ( _fnDataSource( settings ) == 'ssp' ) {
			// In server-side processing mode, most options are irrelevant since
			// rows not shown don't exist and the index order is the applied order
			// Removed is a special case - for consistency just return an empty
			// array
			return search === 'removed' ?
				[] :
				_range( 0, displayMaster.length );
		}
		else if ( page == 'current' ) {
			// Current page implies that order=current and filter=applied, since it is
			// fairly senseless otherwise, regardless of what order and search actually
			// are
			for ( i=settings._iDisplayStart, ien=settings.fnDisplayEnd() ; i<ien ; i++ ) {
				a.push( displayFiltered[i] );
			}
		}
		else if ( order == 'current' || order == 'applied' ) {
			if ( search == 'none') {
				a = displayMaster.slice();
			}
			else if ( search == 'applied' ) {
				a = displayFiltered.slice();
			}
			else if ( search == 'removed' ) {
				// O(n+m) solution by creating a hash map
				var displayFilteredMap = {};
	
				for ( var i=0, ien=displayFiltered.length ; i<ien ; i++ ) {
					displayFilteredMap[displayFiltered[i]] = null;
				}
	
				a = $.map( displayMaster, function (el) {
					return ! displayFilteredMap.hasOwnProperty(el) ?
						el :
						null;
				} );
			}
		}
		else if ( order == 'index' || order == 'original' ) {
			for ( i=0, ien=settings.aoData.length ; i<ien ; i++ ) {
				if ( search == 'none' ) {
					a.push( i );
				}
				else { // applied | removed
					tmp = $.inArray( i, displayFiltered );
	
					if ((tmp === -1 && search == 'removed') ||
						(tmp >= 0   && search == 'applied') )
					{
						a.push( i );
					}
				}
			}
		}
	
		return a;
	};
	
	
	/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
	 * Rows
	 *
	 * {}          - no selector - use all available rows
	 * {integer}   - row aoData index
	 * {node}      - TR node
	 * {string}    - jQuery selector to apply to the TR elements
	 * {array}     - jQuery array of nodes, or simply an array of TR nodes
	 *
	 */
	var __row_selector = function ( settings, selector, opts )
	{
		var rows;
		var run = function ( sel ) {
			var selInt = _intVal( sel );
			var i, ien;
			var aoData = settings.aoData;
	
			// Short cut - selector is a number and no options provided (default is
			// all records, so no need to check if the index is in there, since it
			// must be - dev error if the index doesn't exist).
			if ( selInt !== null && ! opts ) {
				return [ selInt ];
			}
	
			if ( ! rows ) {
				rows = _selector_row_indexes( settings, opts );
			}
	
			if ( selInt !== null && $.inArray( selInt, rows ) !== -1 ) {
				// Selector - integer
				return [ selInt ];
			}
			else if ( sel === null || sel === undefined || sel === '' ) {
				// Selector - none
				return rows;
			}
	
			// Selector - function
			if ( typeof sel === 'function' ) {
				return $.map( rows, function (idx) {
					var row = aoData[ idx ];
					return sel( idx, row._aData, row.nTr ) ? idx : null;
				} );
			}
	
			// Selector - node
			if ( sel.nodeName ) {
				var rowIdx = sel._DT_RowIndex;  // Property added by DT for fast lookup
				var cellIdx = sel._DT_CellIndex;
	
				if ( rowIdx !== undefined ) {
					// Make sure that the row is actually still present in the table
					return aoData[ rowIdx ] && aoData[ rowIdx ].nTr === sel ?
						[ rowIdx ] :
						[];
				}
				else if ( cellIdx ) {
					return aoData[ cellIdx.row ] && aoData[ cellIdx.row ].nTr === sel.parentNode ?
						[ cellIdx.row ] :
						[];
				}
				else {
					var host = $(sel).closest('*[data-dt-row]');
					return host.length ?
						[ host.data('dt-row') ] :
						[];
				}
			}
	
			// ID selector. Want to always be able to select rows by id, regardless
			// of if the tr element has been created or not, so can't rely upon
			// jQuery here - hence a custom implementation. This does not match
			// Sizzle's fast selector or HTML4 - in HTML5 the ID can be anything,
			// but to select it using a CSS selector engine (like Sizzle or
			// querySelect) it would need to need to be escaped for some characters.
			// DataTables simplifies this for row selectors since you can select
			// only a row. A # indicates an id any anything that follows is the id -
			// unescaped.
			if ( typeof sel === 'string' && sel.charAt(0) === '#' ) {
				// get row index from id
				var rowObj = settings.aIds[ sel.replace( /^#/, '' ) ];
				if ( rowObj !== undefined ) {
					return [ rowObj.idx ];
				}
	
				// need to fall through to jQuery in case there is DOM id that
				// matches
			}
			
			// Get nodes in the order from the `rows` array with null values removed
			var nodes = _removeEmpty(
				_pluck_order( settings.aoData, rows, 'nTr' )
			);
	
			// Selector - jQuery selector string, array of nodes or jQuery object/
			// As jQuery's .filter() allows jQuery objects to be passed in filter,
			// it also allows arrays, so this will cope with all three options
			return $(nodes)
				.filter( sel )
				.map( function () {
					return this._DT_RowIndex;
				} )
				.toArray();
		};
	
		return _selector_run( 'row', selector, run, settings, opts );
	};
	
	
	_api_register( 'rows()', function ( selector, opts ) {
		// argument shifting
		if ( selector === undefined ) {
			selector = '';
		}
		else if ( $.isPlainObject( selector ) ) {
			opts = selector;
			selector = '';
		}
	
		opts = _selector_opts( opts );
	
		var inst = this.iterator( 'table', function ( settings ) {
			return __row_selector( settings, selector, opts );
		}, 1 );
	
		// Want argument shifting here and in __row_selector?
		inst.selector.rows = selector;
		inst.selector.opts = opts;
	
		return inst;
	} );
	
	_api_register( 'rows().nodes()', function () {
		return this.iterator( 'row', function ( settings, row ) {
			return settings.aoData[ row ].nTr || undefined;
		}, 1 );
	} );
	
	_api_register( 'rows().data()', function () {
		return this.iterator( true, 'rows', function ( settings, rows ) {
			return _pluck_order( settings.aoData, rows, '_aData' );
		}, 1 );
	} );
	
	_api_registerPlural( 'rows().cache()', 'row().cache()', function ( type ) {
		return this.iterator( 'row', function ( settings, row ) {
			var r = settings.aoData[ row ];
			return type === 'search' ? r._aFilterData : r._aSortData;
		}, 1 );
	} );
	
	_api_registerPlural( 'rows().invalidate()', 'row().invalidate()', function ( src ) {
		return this.iterator( 'row', function ( settings, row ) {
			_fnInvalidate( settings, row, src );
		} );
	} );
	
	_api_registerPlural( 'rows().indexes()', 'row().index()', function () {
		return this.iterator( 'row', function ( settings, row ) {
			return row;
		}, 1 );
	} );
	
	_api_registerPlural( 'rows().ids()', 'row().id()', function ( hash ) {
		var a = [];
		var context = this.context;
	
		// `iterator` will drop undefined values, but in this case we want them
		for ( var i=0, ien=context.length ; i<ien ; i++ ) {
			for ( var j=0, jen=this[i].length ; j<jen ; j++ ) {
				var id = context[i].rowIdFn( context[i].aoData[ this[i][j] ]._aData );
				a.push( (hash === true ? '#' : '' )+ id );
			}
		}
	
		return new _Api( context, a );
	} );
	
	_api_registerPlural( 'rows().remove()', 'row().remove()', function () {
		var that = this;
	
		this.iterator( 'row', function ( settings, row, thatIdx ) {
			var data = settings.aoData;
			var rowData = data[ row ];
			var i, ien, j, jen;
			var loopRow, loopCells;
	
			data.splice( row, 1 );
	
			// Update the cached indexes
			for ( i=0, ien=data.length ; i<ien ; i++ ) {
				loopRow = data[i];
				loopCells = loopRow.anCells;
	
				// Rows
				if ( loopRow.nTr !== null ) {
					loopRow.nTr._DT_RowIndex = i;
				}
	
				// Cells
				if ( loopCells !== null ) {
					for ( j=0, jen=loopCells.length ; j<jen ; j++ ) {
						loopCells[j]._DT_CellIndex.row = i;
					}
				}
			}
	
			// Delete from the display arrays
			_fnDeleteIndex( settings.aiDisplayMaster, row );
			_fnDeleteIndex( settings.aiDisplay, row );
			_fnDeleteIndex( that[ thatIdx ], row, false ); // maintain local indexes
	
			// For server-side processing tables - subtract the deleted row from the count
			if ( settings._iRecordsDisplay > 0 ) {
				settings._iRecordsDisplay--;
			}
	
			// Check for an 'overflow' they case for displaying the table
			_fnLengthOverflow( settings );
	
			// Remove the row's ID reference if there is one
			var id = settings.rowIdFn( rowData._aData );
			if ( id !== undefined ) {
				delete settings.aIds[ id ];
			}
		} );
	
		this.iterator( 'table', function ( settings ) {
			for ( var i=0, ien=settings.aoData.length ; i<ien ; i++ ) {
				settings.aoData[i].idx = i;
			}
		} );
	
		return this;
	} );
	
	
	_api_register( 'rows.add()', function ( rows ) {
		var newRows = this.iterator( 'table', function ( settings ) {
				var row, i, ien;
				var out = [];
	
				for ( i=0, ien=rows.length ; i<ien ; i++ ) {
					row = rows[i];
	
					if ( row.nodeName && row.nodeName.toUpperCase() === 'TR' ) {
						out.push( _fnAddTr( settings, row )[0] );
					}
					else {
						out.push( _fnAddData( settings, row ) );
					}
				}
	
				return out;
			}, 1 );
	
		// Return an Api.rows() extended instance, so rows().nodes() etc can be used
		var modRows = this.rows( -1 );
		modRows.pop();
		$.merge( modRows, newRows );
	
		return modRows;
	} );
	
	
	
	
	
	/**
	 *
	 */
	_api_register( 'row()', function ( selector, opts ) {
		return _selector_first( this.rows( selector, opts ) );
	} );
	
	
	_api_register( 'row().data()', function ( data ) {
		var ctx = this.context;
	
		if ( data === undefined ) {
			// Get
			return ctx.length && this.length ?
				ctx[0].aoData[ this[0] ]._aData :
				undefined;
		}
	
		// Set
		var row = ctx[0].aoData[ this[0] ];
		row._aData = data;
	
		// If the DOM has an id, and the data source is an array
		if ( Array.isArray( data ) && row.nTr && row.nTr.id ) {
			_fnSetObjectDataFn( ctx[0].rowId )( data, row.nTr.id );
		}
	
		// Automatically invalidate
		_fnInvalidate( ctx[0], this[0], 'data' );
	
		return this;
	} );
	
	
	_api_register( 'row().node()', function () {
		var ctx = this.context;
	
		return ctx.length && this.length ?
			ctx[0].aoData[ this[0] ].nTr || null :
			null;
	} );
	
	
	_api_register( 'row.add()', function ( row ) {
		// Allow a jQuery object to be passed in - only a single row is added from
		// it though - the first element in the set
		if ( row instanceof $ && row.length ) {
			row = row[0];
		}
	
		var rows = this.iterator( 'table', function ( settings ) {
			if ( row.nodeName && row.nodeName.toUpperCase() === 'TR' ) {
				return _fnAddTr( settings, row )[0];
			}
			return _fnAddData( settings, row );
		} );
	
		// Return an Api.rows() extended instance, with the newly added row selected
		return this.row( rows[0] );
	} );
	
	
	$(document).on('plugin-init.dt', function (e, context) {
		var api = new _Api( context );
		var namespace = 'on-plugin-init';
		var stateSaveParamsEvent = 'stateSaveParams.' + namespace;
		var destroyEvent = 'destroy. ' + namespace;
	
		api.on( stateSaveParamsEvent, function ( e, settings, d ) {
			// This could be more compact with the API, but it is a lot faster as a simple
			// internal loop
			var idFn = settings.rowIdFn;
			var data = settings.aoData;
			var ids = [];
	
			for (var i=0 ; i<data.length ; i++) {
				if (data[i]._detailsShow) {
					ids.push( '#' + idFn(data[i]._aData) );
				}
			}
	
			d.childRows = ids;
		});
	
		api.on( destroyEvent, function () {
			api.off(stateSaveParamsEvent + ' ' + destroyEvent);
		});
	
		var loaded = api.state.loaded();
	
		if ( loaded && loaded.childRows ) {
			api
				.rows( $.map(loaded.childRows, function (id){
					return id.replace(/:/g, '\\:')
				}) )
				.every( function () {
					_fnCallbackFire( context, null, 'requestChild', [ this ] )
				});
		}
	});
	
	var __details_add = function ( ctx, row, data, klass )
	{
		// Convert to array of TR elements
		var rows = [];
		var addRow = function ( r, k ) {
			// Recursion to allow for arrays of jQuery objects
			if ( Array.isArray( r ) || r instanceof $ ) {
				for ( var i=0, ien=r.length ; i<ien ; i++ ) {
					addRow( r[i], k );
				}
				return;
			}
	
			// If we get a TR element, then just add it directly - up to the dev
			// to add the correct number of columns etc
			if ( r.nodeName && r.nodeName.toLowerCase() === 'tr' ) {
				rows.push( r );
			}
			else {
				// Otherwise create a row with a wrapper
				var created = $('<tr><td></td></tr>').addClass( k );
				$('td', created)
					.addClass( k )
					.html( r )
					[0].colSpan = _fnVisbleColumns( ctx );
	
				rows.push( created[0] );
			}
		};
	
		addRow( data, klass );
	
		if ( row._details ) {
			row._details.detach();
		}
	
		row._details = $(rows);
	
		// If the children were already shown, that state should be retained
		if ( row._detailsShow ) {
			row._details.insertAfter( row.nTr );
		}
	};
	
	
	// Make state saving of child row details async to allow them to be batch processed
	var __details_state = DataTable.util.throttle(
		function (ctx) {
			_fnSaveState( ctx[0] )
		},
		500
	);
	
	
	var __details_remove = function ( api, idx )
	{
		var ctx = api.context;
	
		if ( ctx.length ) {
			var row = ctx[0].aoData[ idx !== undefined ? idx : api[0] ];
	
			if ( row && row._details ) {
				row._details.remove();
	
				row._detailsShow = undefined;
				row._details = undefined;
				$( row.nTr ).removeClass( 'dt-hasChild' );
				__details_state( ctx );
			}
		}
	};
	
	
	var __details_display = function ( api, show ) {
		var ctx = api.context;
	
		if ( ctx.length && api.length ) {
			var row = ctx[0].aoData[ api[0] ];
	
			if ( row._details ) {
				row._detailsShow = show;
	
				if ( show ) {
					row._details.insertAfter( row.nTr );
					$( row.nTr ).addClass( 'dt-hasChild' );
				}
				else {
					row._details.detach();
					$( row.nTr ).removeClass( 'dt-hasChild' );
				}
	
				_fnCallbackFire( ctx[0], null, 'childRow', [ show, api.row( api[0] ) ] )
	
				__details_events( ctx[0] );
				__details_state( ctx );
			}
		}
	};
	
	
	var __details_events = function ( settings )
	{
		var api = new _Api( settings );
		var namespace = '.dt.DT_details';
		var drawEvent = 'draw'+namespace;
		var colvisEvent = 'column-sizing'+namespace;
		var destroyEvent = 'destroy'+namespace;
		var data = settings.aoData;
	
		api.off( drawEvent +' '+ colvisEvent +' '+ destroyEvent );
	
		if ( _pluck( data, '_details' ).length > 0 ) {
			// On each draw, insert the required elements into the document
			api.on( drawEvent, function ( e, ctx ) {
				if ( settings !== ctx ) {
					return;
				}
	
				api.rows( {page:'current'} ).eq(0).each( function (idx) {
					// Internal data grab
					var row = data[ idx ];
	
					if ( row._detailsShow ) {
						row._details.insertAfter( row.nTr );
					}
				} );
			} );
	
			// Column visibility change - update the colspan
			api.on( colvisEvent, function ( e, ctx, idx, vis ) {
				if ( settings !== ctx ) {
					return;
				}
	
				// Update the colspan for the details rows (note, only if it already has
				// a colspan)
				var row, visible = _fnVisbleColumns( ctx );
	
				for ( var i=0, ien=data.length ; i<ien ; i++ ) {
					row = data[i];
	
					if ( row._details ) {
						row._details.children('td[colspan]').attr('colspan', visible );
					}
				}
			} );
	
			// Table destroyed - nuke any child rows
			api.on( destroyEvent, function ( e, ctx ) {
				if ( settings !== ctx ) {
					return;
				}
	
				for ( var i=0, ien=data.length ; i<ien ; i++ ) {
					if ( data[i]._details ) {
						__details_remove( api, i );
					}
				}
			} );
		}
	};
	
	// Strings for the method names to help minification
	var _emp = '';
	var _child_obj = _emp+'row().child';
	var _child_mth = _child_obj+'()';
	
	// data can be:
	//  tr
	//  string
	//  jQuery or array of any of the above
	_api_register( _child_mth, function ( data, klass ) {
		var ctx = this.context;
	
		if ( data === undefined ) {
			// get
			return ctx.length && this.length ?
				ctx[0].aoData[ this[0] ]._details :
				undefined;
		}
		else if ( data === true ) {
			// show
			this.child.show();
		}
		else if ( data === false ) {
			// remove
			__details_remove( this );
		}
		else if ( ctx.length && this.length ) {
			// set
			__details_add( ctx[0], ctx[0].aoData[ this[0] ], data, klass );
		}
	
		return this;
	} );
	
	
	_api_register( [
		_child_obj+'.show()',
		_child_mth+'.show()' // only when `child()` was called with parameters (without
	], function ( show ) {   // it returns an object and this method is not executed)
		__details_display( this, true );
		return this;
	} );
	
	
	_api_register( [
		_child_obj+'.hide()',
		_child_mth+'.hide()' // only when `child()` was called with parameters (without
	], function () {         // it returns an object and this method is not executed)
		__details_display( this, false );
		return this;
	} );
	
	
	_api_register( [
		_child_obj+'.remove()',
		_child_mth+'.remove()' // only when `child()` was called with parameters (without
	], function () {           // it returns an object and this method is not executed)
		__details_remove( this );
		return this;
	} );
	
	
	_api_register( _child_obj+'.isShown()', function () {
		var ctx = this.context;
	
		if ( ctx.length && this.length ) {
			// _detailsShown as false or undefined will fall through to return false
			return ctx[0].aoData[ this[0] ]._detailsShow || false;
		}
		return false;
	} );
	
	
	
	/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
	 * Columns
	 *
	 * {integer}           - column index (>=0 count from left, <0 count from right)
	 * "{integer}:visIdx"  - visible column index (i.e. translate to column index)  (>=0 count from left, <0 count from right)
	 * "{integer}:visible" - alias for {integer}:visIdx  (>=0 count from left, <0 count from right)
	 * "{string}:name"     - column name
	 * "{string}"          - jQuery selector on column header nodes
	 *
	 */
	
	// can be an array of these items, comma separated list, or an array of comma
	// separated lists
	
	var __re_column_selector = /^([^:]+):(name|visIdx|visible)$/;
	
	
	// r1 and r2 are redundant - but it means that the parameters match for the
	// iterator callback in columns().data()
	var __columnData = function ( settings, column, r1, r2, rows ) {
		var a = [];
		for ( var row=0, ien=rows.length ; row<ien ; row++ ) {
			a.push( _fnGetCellData( settings, rows[row], column ) );
		}
		return a;
	};
	
	
	var __column_selector = function ( settings, selector, opts )
	{
		var
			columns = settings.aoColumns,
			names = _pluck( columns, 'sName' ),
			nodes = _pluck( columns, 'nTh' );
	
		var run = function ( s ) {
			var selInt = _intVal( s );
	
			// Selector - all
			if ( s === '' ) {
				return _range( columns.length );
			}
	
			// Selector - index
			if ( selInt !== null ) {
				return [ selInt >= 0 ?
					selInt : // Count from left
					columns.length + selInt // Count from right (+ because its a negative value)
				];
			}
	
			// Selector = function
			if ( typeof s === 'function' ) {
				var rows = _selector_row_indexes( settings, opts );
	
				return $.map( columns, function (col, idx) {
					return s(
							idx,
							__columnData( settings, idx, 0, 0, rows ),
							nodes[ idx ]
						) ? idx : null;
				} );
			}
	
			// jQuery or string selector
			var match = typeof s === 'string' ?
				s.match( __re_column_selector ) :
				'';
	
			if ( match ) {
				switch( match[2] ) {
					case 'visIdx':
					case 'visible':
						var idx = parseInt( match[1], 10 );
						// Visible index given, convert to column index
						if ( idx < 0 ) {
							// Counting from the right
							var visColumns = $.map( columns, function (col,i) {
								return col.bVisible ? i : null;
							} );
							return [ visColumns[ visColumns.length + idx ] ];
						}
						// Counting from the left
						return [ _fnVisibleToColumnIndex( settings, idx ) ];
	
					case 'name':
						// match by name. `names` is column index complete and in order
						return $.map( names, function (name, i) {
							return name === match[1] ? i : null;
						} );
	
					default:
						return [];
				}
			}
	
			// Cell in the table body
			if ( s.nodeName && s._DT_CellIndex ) {
				return [ s._DT_CellIndex.column ];
			}
	
			// jQuery selector on the TH elements for the columns
			var jqResult = $( nodes )
				.filter( s )
				.map( function () {
					return $.inArray( this, nodes ); // `nodes` is column index complete and in order
				} )
				.toArray();
	
			if ( jqResult.length || ! s.nodeName ) {
				return jqResult;
			}
	
			// Otherwise a node which might have a `dt-column` data attribute, or be
			// a child or such an element
			var host = $(s).closest('*[data-dt-column]');
			return host.length ?
				[ host.data('dt-column') ] :
				[];
		};
	
		return _selector_run( 'column', selector, run, settings, opts );
	};
	
	
	var __setColumnVis = function ( settings, column, vis ) {
		var
			cols = settings.aoColumns,
			col  = cols[ column ],
			data = settings.aoData,
			row, cells, i, ien, tr;
	
		// Get
		if ( vis === undefined ) {
			return col.bVisible;
		}
	
		// Set
		// No change
		if ( col.bVisible === vis ) {
			return;
		}
	
		if ( vis ) {
			// Insert column
			// Need to decide if we should use appendChild or insertBefore
			var insertBefore = $.inArray( true, _pluck(cols, 'bVisible'), column+1 );
	
			for ( i=0, ien=data.length ; i<ien ; i++ ) {
				tr = data[i].nTr;
				cells = data[i].anCells;
	
				if ( tr ) {
					// insertBefore can act like appendChild if 2nd arg is null
					tr.insertBefore( cells[ column ], cells[ insertBefore ] || null );
				}
			}
		}
		else {
			// Remove column
			$( _pluck( settings.aoData, 'anCells', column ) ).detach();
		}
	
		// Common actions
		col.bVisible = vis;
	};
	
	
	_api_register( 'columns()', function ( selector, opts ) {
		// argument shifting
		if ( selector === undefined ) {
			selector = '';
		}
		else if ( $.isPlainObject( selector ) ) {
			opts = selector;
			selector = '';
		}
	
		opts = _selector_opts( opts );
	
		var inst = this.iterator( 'table', function ( settings ) {
			return __column_selector( settings, selector, opts );
		}, 1 );
	
		// Want argument shifting here and in _row_selector?
		inst.selector.cols = selector;
		inst.selector.opts = opts;
	
		return inst;
	} );
	
	_api_registerPlural( 'columns().header()', 'column().header()', function ( selector, opts ) {
		return this.iterator( 'column', function ( settings, column ) {
			return settings.aoColumns[column].nTh;
		}, 1 );
	} );
	
	_api_registerPlural( 'columns().footer()', 'column().footer()', function ( selector, opts ) {
		return this.iterator( 'column', function ( settings, column ) {
			return settings.aoColumns[column].nTf;
		}, 1 );
	} );
	
	_api_registerPlural( 'columns().data()', 'column().data()', function () {
		return this.iterator( 'column-rows', __columnData, 1 );
	} );
	
	_api_registerPlural( 'columns().dataSrc()', 'column().dataSrc()', function () {
		return this.iterator( 'column', function ( settings, column ) {
			return settings.aoColumns[column].mData;
		}, 1 );
	} );
	
	_api_registerPlural( 'columns().cache()', 'column().cache()', function ( type ) {
		return this.iterator( 'column-rows', function ( settings, column, i, j, rows ) {
			return _pluck_order( settings.aoData, rows,
				type === 'search' ? '_aFilterData' : '_aSortData', column
			);
		}, 1 );
	} );
	
	_api_registerPlural( 'columns().nodes()', 'column().nodes()', function () {
		return this.iterator( 'column-rows', function ( settings, column, i, j, rows ) {
			return _pluck_order( settings.aoData, rows, 'anCells', column ) ;
		}, 1 );
	} );
	
	_api_registerPlural( 'columns().visible()', 'column().visible()', function ( vis, calc ) {
		var that = this;
		var ret = this.iterator( 'column', function ( settings, column ) {
			if ( vis === undefined ) {
				return settings.aoColumns[ column ].bVisible;
			} // else
			__setColumnVis( settings, column, vis );
		} );
	
		// Group the column visibility changes
		if ( vis !== undefined ) {
			this.iterator( 'table', function ( settings ) {
				// Redraw the header after changes
				_fnDrawHead( settings, settings.aoHeader );
				_fnDrawHead( settings, settings.aoFooter );
		
				// Update colspan for no records display. Child rows and extensions will use their own
				// listeners to do this - only need to update the empty table item here
				if ( ! settings.aiDisplay.length ) {
					$(settings.nTBody).find('td[colspan]').attr('colspan', _fnVisbleColumns(settings));
				}
		
				_fnSaveState( settings );
	
				// Second loop once the first is done for events
				that.iterator( 'column', function ( settings, column ) {
					_fnCallbackFire( settings, null, 'column-visibility', [settings, column, vis, calc] );
				} );
	
				if ( calc === undefined || calc ) {
					that.columns.adjust();
				}
			});
		}
	
		return ret;
	} );
	
	_api_registerPlural( 'columns().indexes()', 'column().index()', function ( type ) {
		return this.iterator( 'column', function ( settings, column ) {
			return type === 'visible' ?
				_fnColumnIndexToVisible( settings, column ) :
				column;
		}, 1 );
	} );
	
	_api_register( 'columns.adjust()', function () {
		return this.iterator( 'table', function ( settings ) {
			_fnAdjustColumnSizing( settings );
		}, 1 );
	} );
	
	_api_register( 'column.index()', function ( type, idx ) {
		if ( this.context.length !== 0 ) {
			var ctx = this.context[0];
	
			if ( type === 'fromVisible' || type === 'toData' ) {
				return _fnVisibleToColumnIndex( ctx, idx );
			}
			else if ( type === 'fromData' || type === 'toVisible' ) {
				return _fnColumnIndexToVisible( ctx, idx );
			}
		}
	} );
	
	_api_register( 'column()', function ( selector, opts ) {
		return _selector_first( this.columns( selector, opts ) );
	} );
	
	var __cell_selector = function ( settings, selector, opts )
	{
		var data = settings.aoData;
		var rows = _selector_row_indexes( settings, opts );
		var cells = _removeEmpty( _pluck_order( data, rows, 'anCells' ) );
		var allCells = $(_flatten( [], cells ));
		var row;
		var columns = settings.aoColumns.length;
		var a, i, ien, j, o, host;
	
		var run = function ( s ) {
			var fnSelector = typeof s === 'function';
	
			if ( s === null || s === undefined || fnSelector ) {
				// All cells and function selectors
				a = [];
	
				for ( i=0, ien=rows.length ; i<ien ; i++ ) {
					row = rows[i];
	
					for ( j=0 ; j<columns ; j++ ) {
						o = {
							row: row,
							column: j
						};
	
						if ( fnSelector ) {
							// Selector - function
							host = data[ row ];
	
							if ( s( o, _fnGetCellData(settings, row, j), host.anCells ? host.anCells[j] : null ) ) {
								a.push( o );
							}
						}
						else {
							// Selector - all
							a.push( o );
						}
					}
				}
	
				return a;
			}
			
			// Selector - index
			if ( $.isPlainObject( s ) ) {
				// Valid cell index and its in the array of selectable rows
				return s.column !== undefined && s.row !== undefined && $.inArray( s.row, rows ) !== -1 ?
					[s] :
					[];
			}
	
			// Selector - jQuery filtered cells
			var jqResult = allCells
				.filter( s )
				.map( function (i, el) {
					return { // use a new object, in case someone changes the values
						row:    el._DT_CellIndex.row,
						column: el._DT_CellIndex.column
	 				};
				} )
				.toArray();
	
			if ( jqResult.length || ! s.nodeName ) {
				return jqResult;
			}
	
			// Otherwise the selector is a node, and there is one last option - the
			// element might be a child of an element which has dt-row and dt-column
			// data attributes
			host = $(s).closest('*[data-dt-row]');
			return host.length ?
				[ {
					row: host.data('dt-row'),
					column: host.data('dt-column')
				} ] :
				[];
		};
	
		return _selector_run( 'cell', selector, run, settings, opts );
	};
	
	
	
	
	_api_register( 'cells()', function ( rowSelector, columnSelector, opts ) {
		// Argument shifting
		if ( $.isPlainObject( rowSelector ) ) {
			// Indexes
			if ( rowSelector.row === undefined ) {
				// Selector options in first parameter
				opts = rowSelector;
				rowSelector = null;
			}
			else {
				// Cell index objects in first parameter
				opts = columnSelector;
				columnSelector = null;
			}
		}
		if ( $.isPlainObject( columnSelector ) ) {
			opts = columnSelector;
			columnSelector = null;
		}
	
		// Cell selector
		if ( columnSelector === null || columnSelector === undefined ) {
			return this.iterator( 'table', function ( settings ) {
				return __cell_selector( settings, rowSelector, _selector_opts( opts ) );
			} );
		}
	
		// The default built in options need to apply to row and columns
		var internalOpts = opts ? {
			page: opts.page,
			order: opts.order,
			search: opts.search
		} : {};
	
		// Row + column selector
		var columns = this.columns( columnSelector, internalOpts );
		var rows = this.rows( rowSelector, internalOpts );
		var i, ien, j, jen;
	
		var cellsNoOpts = this.iterator( 'table', function ( settings, idx ) {
			var a = [];
	
			for ( i=0, ien=rows[idx].length ; i<ien ; i++ ) {
				for ( j=0, jen=columns[idx].length ; j<jen ; j++ ) {
					a.push( {
						row:    rows[idx][i],
						column: columns[idx][j]
					} );
				}
			}
	
			return a;
		}, 1 );
	
		// There is currently only one extension which uses a cell selector extension
		// It is a _major_ performance drag to run this if it isn't needed, so this is
		// an extension specific check at the moment
		var cells = opts && opts.selected ?
			this.cells( cellsNoOpts, opts ) :
			cellsNoOpts;
	
		$.extend( cells.selector, {
			cols: columnSelector,
			rows: rowSelector,
			opts: opts
		} );
	
		return cells;
	} );
	
	
	_api_registerPlural( 'cells().nodes()', 'cell().node()', function () {
		return this.iterator( 'cell', function ( settings, row, column ) {
			var data = settings.aoData[ row ];
	
			return data && data.anCells ?
				data.anCells[ column ] :
				undefined;
		}, 1 );
	} );
	
	
	_api_register( 'cells().data()', function () {
		return this.iterator( 'cell', function ( settings, row, column ) {
			return _fnGetCellData( settings, row, column );
		}, 1 );
	} );
	
	
	_api_registerPlural( 'cells().cache()', 'cell().cache()', function ( type ) {
		type = type === 'search' ? '_aFilterData' : '_aSortData';
	
		return this.iterator( 'cell', function ( settings, row, column ) {
			return settings.aoData[ row ][ type ][ column ];
		}, 1 );
	} );
	
	
	_api_registerPlural( 'cells().render()', 'cell().render()', function ( type ) {
		return this.iterator( 'cell', function ( settings, row, column ) {
			return _fnGetCellData( settings, row, column, type );
		}, 1 );
	} );
	
	
	_api_registerPlural( 'cells().indexes()', 'cell().index()', function () {
		return this.iterator( 'cell', function ( settings, row, column ) {
			return {
				row: row,
				column: column,
				columnVisible: _fnColumnIndexToVisible( settings, column )
			};
		}, 1 );
	} );
	
	
	_api_registerPlural( 'cells().invalidate()', 'cell().invalidate()', function ( src ) {
		return this.iterator( 'cell', function ( settings, row, column ) {
			_fnInvalidate( settings, row, src, column );
		} );
	} );
	
	
	
	_api_register( 'cell()', function ( rowSelector, columnSelector, opts ) {
		return _selector_first( this.cells( rowSelector, columnSelector, opts ) );
	} );
	
	
	_api_register( 'cell().data()', function ( data ) {
		var ctx = this.context;
		var cell = this[0];
	
		if ( data === undefined ) {
			// Get
			return ctx.length && cell.length ?
				_fnGetCellData( ctx[0], cell[0].row, cell[0].column ) :
				undefined;
		}
	
		// Set
		_fnSetCellData( ctx[0], cell[0].row, cell[0].column, data );
		_fnInvalidate( ctx[0], cell[0].row, 'data', cell[0].column );
	
		return this;
	} );
	
	
	
	/**
	 * Get current ordering (sorting) that has been applied to the table.
	 *
	 * @returns {array} 2D array containing the sorting information for the first
	 *   table in the current context. Each element in the parent array represents
	 *   a column being sorted upon (i.e. multi-sorting with two columns would have
	 *   2 inner arrays). The inner arrays may have 2 or 3 elements. The first is
	 *   the column index that the sorting condition applies to, the second is the
	 *   direction of the sort (`desc` or `asc`) and, optionally, the third is the
	 *   index of the sorting order from the `column.sorting` initialisation array.
	 *//**
	 * Set the ordering for the table.
	 *
	 * @param {integer} order Column index to sort upon.
	 * @param {string} direction Direction of the sort to be applied (`asc` or `desc`)
	 * @returns {DataTables.Api} this
	 *//**
	 * Set the ordering for the table.
	 *
	 * @param {array} order 1D array of sorting information to be applied.
	 * @param {array} [...] Optional additional sorting conditions
	 * @returns {DataTables.Api} this
	 *//**
	 * Set the ordering for the table.
	 *
	 * @param {array} order 2D array of sorting information to be applied.
	 * @returns {DataTables.Api} this
	 */
	_api_register( 'order()', function ( order, dir ) {
		var ctx = this.context;
	
		if ( order === undefined ) {
			// get
			return ctx.length !== 0 ?
				ctx[0].aaSorting :
				undefined;
		}
	
		// set
		if ( typeof order === 'number' ) {
			// Simple column / direction passed in
			order = [ [ order, dir ] ];
		}
		else if ( order.length && ! Array.isArray( order[0] ) ) {
			// Arguments passed in (list of 1D arrays)
			order = Array.prototype.slice.call( arguments );
		}
		// otherwise a 2D array was passed in
	
		return this.iterator( 'table', function ( settings ) {
			settings.aaSorting = order.slice();
		} );
	} );
	
	
	/**
	 * Attach a sort listener to an element for a given column
	 *
	 * @param {node|jQuery|string} node Identifier for the element(s) to attach the
	 *   listener to. This can take the form of a single DOM node, a jQuery
	 *   collection of nodes or a jQuery selector which will identify the node(s).
	 * @param {integer} column the column that a click on this node will sort on
	 * @param {function} [callback] callback function when sort is run
	 * @returns {DataTables.Api} this
	 */
	_api_register( 'order.listener()', function ( node, column, callback ) {
		return this.iterator( 'table', function ( settings ) {
			_fnSortAttachListener( settings, node, column, callback );
		} );
	} );
	
	
	_api_register( 'order.fixed()', function ( set ) {
		if ( ! set ) {
			var ctx = this.context;
			var fixed = ctx.length ?
				ctx[0].aaSortingFixed :
				undefined;
	
			return Array.isArray( fixed ) ?
				{ pre: fixed } :
				fixed;
		}
	
		return this.iterator( 'table', function ( settings ) {
			settings.aaSortingFixed = $.extend( true, {}, set );
		} );
	} );
	
	
	// Order by the selected column(s)
	_api_register( [
		'columns().order()',
		'column().order()'
	], function ( dir ) {
		var that = this;
	
		return this.iterator( 'table', function ( settings, i ) {
			var sort = [];
	
			$.each( that[i], function (j, col) {
				sort.push( [ col, dir ] );
			} );
	
			settings.aaSorting = sort;
		} );
	} );
	
	
	
	_api_register( 'search()', function ( input, regex, smart, caseInsen ) {
		var ctx = this.context;
	
		if ( input === undefined ) {
			// get
			return ctx.length !== 0 ?
				ctx[0].oPreviousSearch.sSearch :
				undefined;
		}
	
		// set
		return this.iterator( 'table', function ( settings ) {
			if ( ! settings.oFeatures.bFilter ) {
				return;
			}
	
			_fnFilterComplete( settings, $.extend( {}, settings.oPreviousSearch, {
				"sSearch": input+"",
				"bRegex":  regex === null ? false : regex,
				"bSmart":  smart === null ? true  : smart,
				"bCaseInsensitive": caseInsen === null ? true : caseInsen
			} ), 1 );
		} );
	} );
	
	
	_api_registerPlural(
		'columns().search()',
		'column().search()',
		function ( input, regex, smart, caseInsen ) {
			return this.iterator( 'column', function ( settings, column ) {
				var preSearch = settings.aoPreSearchCols;
	
				if ( input === undefined ) {
					// get
					return preSearch[ column ].sSearch;
				}
	
				// set
				if ( ! settings.oFeatures.bFilter ) {
					return;
				}
	
				$.extend( preSearch[ column ], {
					"sSearch": input+"",
					"bRegex":  regex === null ? false : regex,
					"bSmart":  smart === null ? true  : smart,
					"bCaseInsensitive": caseInsen === null ? true : caseInsen
				} );
	
				_fnFilterComplete( settings, settings.oPreviousSearch, 1 );
			} );
		}
	);
	
	/*
	 * State API methods
	 */
	
	_api_register( 'state()', function () {
		return this.context.length ?
			this.context[0].oSavedState :
			null;
	} );
	
	
	_api_register( 'state.clear()', function () {
		return this.iterator( 'table', function ( settings ) {
			// Save an empty object
			settings.fnStateSaveCallback.call( settings.oInstance, settings, {} );
		} );
	} );
	
	
	_api_register( 'state.loaded()', function () {
		return this.context.length ?
			this.context[0].oLoadedState :
			null;
	} );
	
	
	_api_register( 'state.save()', function () {
		return this.iterator( 'table', function ( settings ) {
			_fnSaveState( settings );
		} );
	} );
	
	
	
	/**
	 * Set the jQuery or window object to be used by DataTables
	 *
	 * @param {*} module Library / container object
	 * @param {string} type Library or container type `lib` or `win`.
	 */
	DataTable.use = function (module, type) {
		if (type === 'lib' || module.fn) {
			$ = module;
		}
		else if (type == 'win' || module.document) {
			window = module;
			document = module.document;
		}
	}
	
	/**
	 * CommonJS factory function pass through. This will check if the arguments
	 * given are a window object or a jQuery object. If so they are set
	 * accordingly.
	 * @param {*} root Window
	 * @param {*} jq jQUery
	 * @returns {boolean} Indicator
	 */
	DataTable.factory = function (root, jq) {
		var is = false;
	
		// Test if the first parameter is a window object
		if (root && root.document) {
			window = root;
			document = root.document;
		}
	
		// Test if the second parameter is a jQuery object
		if (jq && jq.fn && jq.fn.jquery) {
			$ = jq;
			is = true;
		}
	
		return is;
	}
	
	/**
	 * Provide a common method for plug-ins to check the version of DataTables being
	 * used, in order to ensure compatibility.
	 *
	 *  @param {string} version Version string to check for, in the format "X.Y.Z".
	 *    Note that the formats "X" and "X.Y" are also acceptable.
	 *  @returns {boolean} true if this version of DataTables is greater or equal to
	 *    the required version, or false if this version of DataTales is not
	 *    suitable
	 *  @static
	 *  @dtopt API-Static
	 *
	 *  @example
	 *    alert( $.fn.dataTable.versionCheck( '1.9.0' ) );
	 */
	DataTable.versionCheck = DataTable.fnVersionCheck = function( version )
	{
		var aThis = DataTable.version.split('.');
		var aThat = version.split('.');
		var iThis, iThat;
	
		for ( var i=0, iLen=aThat.length ; i<iLen ; i++ ) {
			iThis = parseInt( aThis[i], 10 ) || 0;
			iThat = parseInt( aThat[i], 10 ) || 0;
	
			// Parts are the same, keep comparing
			if (iThis === iThat) {
				continue;
			}
	
			// Parts are different, return immediately
			return iThis > iThat;
		}
	
		return true;
	};
	
	
	/**
	 * Check if a `<table>` node is a DataTable table already or not.
	 *
	 *  @param {node|jquery|string} table Table node, jQuery object or jQuery
	 *      selector for the table to test. Note that if more than more than one
	 *      table is passed on, only the first will be checked
	 *  @returns {boolean} true the table given is a DataTable, or false otherwise
	 *  @static
	 *  @dtopt API-Static
	 *
	 *  @example
	 *    if ( ! $.fn.DataTable.isDataTable( '#example' ) ) {
	 *      $('#example').dataTable();
	 *    }
	 */
	DataTable.isDataTable = DataTable.fnIsDataTable = function ( table )
	{
		var t = $(table).get(0);
		var is = false;
	
		if ( table instanceof DataTable.Api ) {
			return true;
		}
	
		$.each( DataTable.settings, function (i, o) {
			var head = o.nScrollHead ? $('table', o.nScrollHead)[0] : null;
			var foot = o.nScrollFoot ? $('table', o.nScrollFoot)[0] : null;
	
			if ( o.nTable === t || head === t || foot === t ) {
				is = true;
			}
		} );
	
		return is;
	};
	
	
	/**
	 * Get all DataTable tables that have been initialised - optionally you can
	 * select to get only currently visible tables.
	 *
	 *  @param {boolean} [visible=false] Flag to indicate if you want all (default)
	 *    or visible tables only.
	 *  @returns {array} Array of `table` nodes (not DataTable instances) which are
	 *    DataTables
	 *  @static
	 *  @dtopt API-Static
	 *
	 *  @example
	 *    $.each( $.fn.dataTable.tables(true), function () {
	 *      $(table).DataTable().columns.adjust();
	 *    } );
	 */
	DataTable.tables = DataTable.fnTables = function ( visible )
	{
		var api = false;
	
		if ( $.isPlainObject( visible ) ) {
			api = visible.api;
			visible = visible.visible;
		}
	
		var a = $.map( DataTable.settings, function (o) {
			if ( !visible || (visible && $(o.nTable).is(':visible')) ) {
				return o.nTable;
			}
		} );
	
		return api ?
			new _Api( a ) :
			a;
	};
	
	
	/**
	 * Convert from camel case parameters to Hungarian notation. This is made public
	 * for the extensions to provide the same ability as DataTables core to accept
	 * either the 1.9 style Hungarian notation, or the 1.10+ style camelCase
	 * parameters.
	 *
	 *  @param {object} src The model object which holds all parameters that can be
	 *    mapped.
	 *  @param {object} user The object to convert from camel case to Hungarian.
	 *  @param {boolean} force When set to `true`, properties which already have a
	 *    Hungarian value in the `user` object will be overwritten. Otherwise they
	 *    won't be.
	 */
	DataTable.camelToHungarian = _fnCamelToHungarian;
	
	
	
	/**
	 *
	 */
	_api_register( '$()', function ( selector, opts ) {
		var
			rows   = this.rows( opts ).nodes(), // Get all rows
			jqRows = $(rows);
	
		return $( [].concat(
			jqRows.filter( selector ).toArray(),
			jqRows.find( selector ).toArray()
		) );
	} );
	
	
	// jQuery functions to operate on the tables
	$.each( [ 'on', 'one', 'off' ], function (i, key) {
		_api_register( key+'()', function ( /* event, handler */ ) {
			var args = Array.prototype.slice.call(arguments);
	
			// Add the `dt` namespace automatically if it isn't already present
			args[0] = $.map( args[0].split( /\s/ ), function ( e ) {
				return ! e.match(/\.dt\b/) ?
					e+'.dt' :
					e;
				} ).join( ' ' );
	
			var inst = $( this.tables().nodes() );
			inst[key].apply( inst, args );
			return this;
		} );
	} );
	
	
	_api_register( 'clear()', function () {
		return this.iterator( 'table', function ( settings ) {
			_fnClearTable( settings );
		} );
	} );
	
	
	_api_register( 'settings()', function () {
		return new _Api( this.context, this.context );
	} );
	
	
	_api_register( 'init()', function () {
		var ctx = this.context;
		return ctx.length ? ctx[0].oInit : null;
	} );
	
	
	_api_register( 'data()', function () {
		return this.iterator( 'table', function ( settings ) {
			return _pluck( settings.aoData, '_aData' );
		} ).flatten();
	} );
	
	
	_api_register( 'destroy()', function ( remove ) {
		remove = remove || false;
	
		return this.iterator( 'table', function ( settings ) {
			var classes   = settings.oClasses;
			var table     = settings.nTable;
			var tbody     = settings.nTBody;
			var thead     = settings.nTHead;
			var tfoot     = settings.nTFoot;
			var jqTable   = $(table);
			var jqTbody   = $(tbody);
			var jqWrapper = $(settings.nTableWrapper);
			var rows      = $.map( settings.aoData, function (r) { return r.nTr; } );
			var i, ien;
	
			// Flag to note that the table is currently being destroyed - no action
			// should be taken
			settings.bDestroying = true;
	
			// Fire off the destroy callbacks for plug-ins etc
			_fnCallbackFire( settings, "aoDestroyCallback", "destroy", [settings] );
	
			// If not being removed from the document, make all columns visible
			if ( ! remove ) {
				new _Api( settings ).columns().visible( true );
			}
	
			// Blitz all `DT` namespaced events (these are internal events, the
			// lowercase, `dt` events are user subscribed and they are responsible
			// for removing them
			jqWrapper.off('.DT').find(':not(tbody *)').off('.DT');
			$(window).off('.DT-'+settings.sInstance);
	
			// When scrolling we had to break the table up - restore it
			if ( table != thead.parentNode ) {
				jqTable.children('thead').detach();
				jqTable.append( thead );
			}
	
			if ( tfoot && table != tfoot.parentNode ) {
				jqTable.children('tfoot').detach();
				jqTable.append( tfoot );
			}
	
			settings.aaSorting = [];
			settings.aaSortingFixed = [];
			_fnSortingClasses( settings );
	
			$( rows ).removeClass( settings.asStripeClasses.join(' ') );
	
			$('th, td', thead).removeClass( classes.sSortable+' '+
				classes.sSortableAsc+' '+classes.sSortableDesc+' '+classes.sSortableNone
			);
	
			// Add the TR elements back into the table in their original order
			jqTbody.children().detach();
			jqTbody.append( rows );
	
			var orig = settings.nTableWrapper.parentNode;
	
			// Remove the DataTables generated nodes, events and classes
			var removedMethod = remove ? 'remove' : 'detach';
			jqTable[ removedMethod ]();
			jqWrapper[ removedMethod ]();
	
			// If we need to reattach the table to the document
			if ( ! remove && orig ) {
				// insertBefore acts like appendChild if !arg[1]
				orig.insertBefore( table, settings.nTableReinsertBefore );
	
				// Restore the width of the original table - was read from the style property,
				// so we can restore directly to that
				jqTable
					.css( 'width', settings.sDestroyWidth )
					.removeClass( classes.sTable );
	
				// If the were originally stripe classes - then we add them back here.
				// Note this is not fool proof (for example if not all rows had stripe
				// classes - but it's a good effort without getting carried away
				ien = settings.asDestroyStripes.length;
	
				if ( ien ) {
					jqTbody.children().each( function (i) {
						$(this).addClass( settings.asDestroyStripes[i % ien] );
					} );
				}
			}
	
			/* Remove the settings object from the settings array */
			var idx = $.inArray( settings, DataTable.settings );
			if ( idx !== -1 ) {
				DataTable.settings.splice( idx, 1 );
			}
		} );
	} );
	
	
	// Add the `every()` method for rows, columns and cells in a compact form
	$.each( [ 'column', 'row', 'cell' ], function ( i, type ) {
		_api_register( type+'s().every()', function ( fn ) {
			var opts = this.selector.opts;
			var api = this;
	
			return this.iterator( type, function ( settings, arg1, arg2, arg3, arg4 ) {
				// Rows and columns:
				//  arg1 - index
				//  arg2 - table counter
				//  arg3 - loop counter
				//  arg4 - undefined
				// Cells:
				//  arg1 - row index
				//  arg2 - column index
				//  arg3 - table counter
				//  arg4 - loop counter
				fn.call(
					api[ type ](
						arg1,
						type==='cell' ? arg2 : opts,
						type==='cell' ? opts : undefined
					),
					arg1, arg2, arg3, arg4
				);
			} );
		} );
	} );
	
	
	// i18n method for extensions to be able to use the language object from the
	// DataTable
	_api_register( 'i18n()', function ( token, def, plural ) {
		var ctx = this.context[0];
		var resolved = _fnGetObjectDataFn( token )( ctx.oLanguage );
	
		if ( resolved === undefined ) {
			resolved = def;
		}
	
		if ( plural !== undefined && $.isPlainObject( resolved ) ) {
			resolved = resolved[ plural ] !== undefined ?
				resolved[ plural ] :
				resolved._;
		}
	
		return resolved.replace( '%d', plural ); // nb: plural might be undefined,
	} );	
	/**
	 * Version string for plug-ins to check compatibility. Allowed format is
	 * `a.b.c-d` where: a:int, b:int, c:int, d:string(dev|beta|alpha). `d` is used
	 * only for non-release builds. See http://semver.org/ for more information.
	 *  @member
	 *  @type string
	 *  @default Version number
	 */
	DataTable.version = "1.13.4";
	
	/**
	 * Private data store, containing all of the settings objects that are
	 * created for the tables on a given page.
	 *
	 * Note that the `DataTable.settings` object is aliased to
	 * `jQuery.fn.dataTableExt` through which it may be accessed and
	 * manipulated, or `jQuery.fn.dataTable.settings`.
	 *  @member
	 *  @type array
	 *  @default []
	 *  @private
	 */
	DataTable.settings = [];
	
	/**
	 * Object models container, for the various models that DataTables has
	 * available to it. These models define the objects that are used to hold
	 * the active state and configuration of the table.
	 *  @namespace
	 */
	DataTable.models = {};
	
	
	
	/**
	 * Template object for the way in which DataTables holds information about
	 * search information for the global filter and individual column filters.
	 *  @namespace
	 */
	DataTable.models.oSearch = {
		/**
		 * Flag to indicate if the filtering should be case insensitive or not
		 *  @type boolean
		 *  @default true
		 */
		"bCaseInsensitive": true,
	
		/**
		 * Applied search term
		 *  @type string
		 *  @default <i>Empty string</i>
		 */
		"sSearch": "",
	
		/**
		 * Flag to indicate if the search term should be interpreted as a
		 * regular expression (true) or not (false) and therefore and special
		 * regex characters escaped.
		 *  @type boolean
		 *  @default false
		 */
		"bRegex": false,
	
		/**
		 * Flag to indicate if DataTables is to use its smart filtering or not.
		 *  @type boolean
		 *  @default true
		 */
		"bSmart": true,
	
		/**
		 * Flag to indicate if DataTables should only trigger a search when
		 * the return key is pressed.
		 *  @type boolean
		 *  @default false
		 */
		"return": false
	};
	
	
	
	
	/**
	 * Template object for the way in which DataTables holds information about
	 * each individual row. This is the object format used for the settings
	 * aoData array.
	 *  @namespace
	 */
	DataTable.models.oRow = {
		/**
		 * TR element for the row
		 *  @type node
		 *  @default null
		 */
		"nTr": null,
	
		/**
		 * Array of TD elements for each row. This is null until the row has been
		 * created.
		 *  @type array nodes
		 *  @default []
		 */
		"anCells": null,
	
		/**
		 * Data object from the original data source for the row. This is either
		 * an array if using the traditional form of DataTables, or an object if
		 * using mData options. The exact type will depend on the passed in
		 * data from the data source, or will be an array if using DOM a data
		 * source.
		 *  @type array|object
		 *  @default []
		 */
		"_aData": [],
	
		/**
		 * Sorting data cache - this array is ostensibly the same length as the
		 * number of columns (although each index is generated only as it is
		 * needed), and holds the data that is used for sorting each column in the
		 * row. We do this cache generation at the start of the sort in order that
		 * the formatting of the sort data need be done only once for each cell
		 * per sort. This array should not be read from or written to by anything
		 * other than the master sorting methods.
		 *  @type array
		 *  @default null
		 *  @private
		 */
		"_aSortData": null,
	
		/**
		 * Per cell filtering data cache. As per the sort data cache, used to
		 * increase the performance of the filtering in DataTables
		 *  @type array
		 *  @default null
		 *  @private
		 */
		"_aFilterData": null,
	
		/**
		 * Filtering data cache. This is the same as the cell filtering cache, but
		 * in this case a string rather than an array. This is easily computed with
		 * a join on `_aFilterData`, but is provided as a cache so the join isn't
		 * needed on every search (memory traded for performance)
		 *  @type array
		 *  @default null
		 *  @private
		 */
		"_sFilterRow": null,
	
		/**
		 * Cache of the class name that DataTables has applied to the row, so we
		 * can quickly look at this variable rather than needing to do a DOM check
		 * on className for the nTr property.
		 *  @type string
		 *  @default <i>Empty string</i>
		 *  @private
		 */
		"_sRowStripe": "",
	
		/**
		 * Denote if the original data source was from the DOM, or the data source
		 * object. This is used for invalidating data, so DataTables can
		 * automatically read data from the original source, unless uninstructed
		 * otherwise.
		 *  @type string
		 *  @default null
		 *  @private
		 */
		"src": null,
	
		/**
		 * Index in the aoData array. This saves an indexOf lookup when we have the
		 * object, but want to know the index
		 *  @type integer
		 *  @default -1
		 *  @private
		 */
		"idx": -1
	};
	
	
	/**
	 * Template object for the column information object in DataTables. This object
	 * is held in the settings aoColumns array and contains all the information that
	 * DataTables needs about each individual column.
	 *
	 * Note that this object is related to {@link DataTable.defaults.column}
	 * but this one is the internal data store for DataTables's cache of columns.
	 * It should NOT be manipulated outside of DataTables. Any configuration should
	 * be done through the initialisation options.
	 *  @namespace
	 */
	DataTable.models.oColumn = {
		/**
		 * Column index. This could be worked out on-the-fly with $.inArray, but it
		 * is faster to just hold it as a variable
		 *  @type integer
		 *  @default null
		 */
		"idx": null,
	
		/**
		 * A list of the columns that sorting should occur on when this column
		 * is sorted. That this property is an array allows multi-column sorting
		 * to be defined for a column (for example first name / last name columns
		 * would benefit from this). The values are integers pointing to the
		 * columns to be sorted on (typically it will be a single integer pointing
		 * at itself, but that doesn't need to be the case).
		 *  @type array
		 */
		"aDataSort": null,
	
		/**
		 * Define the sorting directions that are applied to the column, in sequence
		 * as the column is repeatedly sorted upon - i.e. the first value is used
		 * as the sorting direction when the column if first sorted (clicked on).
		 * Sort it again (click again) and it will move on to the next index.
		 * Repeat until loop.
		 *  @type array
		 */
		"asSorting": null,
	
		/**
		 * Flag to indicate if the column is searchable, and thus should be included
		 * in the filtering or not.
		 *  @type boolean
		 */
		"bSearchable": null,
	
		/**
		 * Flag to indicate if the column is sortable or not.
		 *  @type boolean
		 */
		"bSortable": null,
	
		/**
		 * Flag to indicate if the column is currently visible in the table or not
		 *  @type boolean
		 */
		"bVisible": null,
	
		/**
		 * Store for manual type assignment using the `column.type` option. This
		 * is held in store so we can manipulate the column's `sType` property.
		 *  @type string
		 *  @default null
		 *  @private
		 */
		"_sManualType": null,
	
		/**
		 * Flag to indicate if HTML5 data attributes should be used as the data
		 * source for filtering or sorting. True is either are.
		 *  @type boolean
		 *  @default false
		 *  @private
		 */
		"_bAttrSrc": false,
	
		/**
		 * Developer definable function that is called whenever a cell is created (Ajax source,
		 * etc) or processed for input (DOM source). This can be used as a compliment to mRender
		 * allowing you to modify the DOM element (add background colour for example) when the
		 * element is available.
		 *  @type function
		 *  @param {element} nTd The TD node that has been created
		 *  @param {*} sData The Data for the cell
		 *  @param {array|object} oData The data for the whole row
		 *  @param {int} iRow The row index for the aoData data store
		 *  @default null
		 */
		"fnCreatedCell": null,
	
		/**
		 * Function to get data from a cell in a column. You should <b>never</b>
		 * access data directly through _aData internally in DataTables - always use
		 * the method attached to this property. It allows mData to function as
		 * required. This function is automatically assigned by the column
		 * initialisation method
		 *  @type function
		 *  @param {array|object} oData The data array/object for the array
		 *    (i.e. aoData[]._aData)
		 *  @param {string} sSpecific The specific data type you want to get -
		 *    'display', 'type' 'filter' 'sort'
		 *  @returns {*} The data for the cell from the given row's data
		 *  @default null
		 */
		"fnGetData": null,
	
		/**
		 * Function to set data for a cell in the column. You should <b>never</b>
		 * set the data directly to _aData internally in DataTables - always use
		 * this method. It allows mData to function as required. This function
		 * is automatically assigned by the column initialisation method
		 *  @type function
		 *  @param {array|object} oData The data array/object for the array
		 *    (i.e. aoData[]._aData)
		 *  @param {*} sValue Value to set
		 *  @default null
		 */
		"fnSetData": null,
	
		/**
		 * Property to read the value for the cells in the column from the data
		 * source array / object. If null, then the default content is used, if a
		 * function is given then the return from the function is used.
		 *  @type function|int|string|null
		 *  @default null
		 */
		"mData": null,
	
		/**
		 * Partner property to mData which is used (only when defined) to get
		 * the data - i.e. it is basically the same as mData, but without the
		 * 'set' option, and also the data fed to it is the result from mData.
		 * This is the rendering method to match the data method of mData.
		 *  @type function|int|string|null
		 *  @default null
		 */
		"mRender": null,
	
		/**
		 * Unique header TH/TD element for this column - this is what the sorting
		 * listener is attached to (if sorting is enabled.)
		 *  @type node
		 *  @default null
		 */
		"nTh": null,
	
		/**
		 * Unique footer TH/TD element for this column (if there is one). Not used
		 * in DataTables as such, but can be used for plug-ins to reference the
		 * footer for each column.
		 *  @type node
		 *  @default null
		 */
		"nTf": null,
	
		/**
		 * The class to apply to all TD elements in the table's TBODY for the column
		 *  @type string
		 *  @default null
		 */
		"sClass": null,
	
		/**
		 * When DataTables calculates the column widths to assign to each column,
		 * it finds the longest string in each column and then constructs a
		 * temporary table and reads the widths from that. The problem with this
		 * is that "mmm" is much wider then "iiii", but the latter is a longer
		 * string - thus the calculation can go wrong (doing it properly and putting
		 * it into an DOM object and measuring that is horribly(!) slow). Thus as
		 * a "work around" we provide this option. It will append its value to the
		 * text that is found to be the longest string for the column - i.e. padding.
		 *  @type string
		 */
		"sContentPadding": null,
	
		/**
		 * Allows a default value to be given for a column's data, and will be used
		 * whenever a null data source is encountered (this can be because mData
		 * is set to null, or because the data source itself is null).
		 *  @type string
		 *  @default null
		 */
		"sDefaultContent": null,
	
		/**
		 * Name for the column, allowing reference to the column by name as well as
		 * by index (needs a lookup to work by name).
		 *  @type string
		 */
		"sName": null,
	
		/**
		 * Custom sorting data type - defines which of the available plug-ins in
		 * afnSortData the custom sorting will use - if any is defined.
		 *  @type string
		 *  @default std
		 */
		"sSortDataType": 'std',
	
		/**
		 * Class to be applied to the header element when sorting on this column
		 *  @type string
		 *  @default null
		 */
		"sSortingClass": null,
	
		/**
		 * Class to be applied to the header element when sorting on this column -
		 * when jQuery UI theming is used.
		 *  @type string
		 *  @default null
		 */
		"sSortingClassJUI": null,
	
		/**
		 * Title of the column - what is seen in the TH element (nTh).
		 *  @type string
		 */
		"sTitle": null,
	
		/**
		 * Column sorting and filtering type
		 *  @type string
		 *  @default null
		 */
		"sType": null,
	
		/**
		 * Width of the column
		 *  @type string
		 *  @default null
		 */
		"sWidth": null,
	
		/**
		 * Width of the column when it was first "encountered"
		 *  @type string
		 *  @default null
		 */
		"sWidthOrig": null
	};
	
	
	/*
	 * Developer note: The properties of the object below are given in Hungarian
	 * notation, that was used as the interface for DataTables prior to v1.10, however
	 * from v1.10 onwards the primary interface is camel case. In order to avoid
	 * breaking backwards compatibility utterly with this change, the Hungarian
	 * version is still, internally the primary interface, but is is not documented
	 * - hence the @name tags in each doc comment. This allows a Javascript function
	 * to create a map from Hungarian notation to camel case (going the other direction
	 * would require each property to be listed, which would add around 3K to the size
	 * of DataTables, while this method is about a 0.5K hit).
	 *
	 * Ultimately this does pave the way for Hungarian notation to be dropped
	 * completely, but that is a massive amount of work and will break current
	 * installs (therefore is on-hold until v2).
	 */
	
	/**
	 * Initialisation options that can be given to DataTables at initialisation
	 * time.
	 *  @namespace
	 */
	DataTable.defaults = {
		/**
		 * An array of data to use for the table, passed in at initialisation which
		 * will be used in preference to any data which is already in the DOM. This is
		 * particularly useful for constructing tables purely in Javascript, for
		 * example with a custom Ajax call.
		 *  @type array
		 *  @default null
		 *
		 *  @dtopt Option
		 *  @name DataTable.defaults.data
		 *
		 *  @example
		 *    // Using a 2D array data source
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "data": [
		 *          ['Trident', 'Internet Explorer 4.0', 'Win 95+', 4, 'X'],
		 *          ['Trident', 'Internet Explorer 5.0', 'Win 95+', 5, 'C'],
		 *        ],
		 *        "columns": [
		 *          { "title": "Engine" },
		 *          { "title": "Browser" },
		 *          { "title": "Platform" },
		 *          { "title": "Version" },
		 *          { "title": "Grade" }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using an array of objects as a data source (`data`)
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "data": [
		 *          {
		 *            "engine":   "Trident",
		 *            "browser":  "Internet Explorer 4.0",
		 *            "platform": "Win 95+",
		 *            "version":  4,
		 *            "grade":    "X"
		 *          },
		 *          {
		 *            "engine":   "Trident",
		 *            "browser":  "Internet Explorer 5.0",
		 *            "platform": "Win 95+",
		 *            "version":  5,
		 *            "grade":    "C"
		 *          }
		 *        ],
		 *        "columns": [
		 *          { "title": "Engine",   "data": "engine" },
		 *          { "title": "Browser",  "data": "browser" },
		 *          { "title": "Platform", "data": "platform" },
		 *          { "title": "Version",  "data": "version" },
		 *          { "title": "Grade",    "data": "grade" }
		 *        ]
		 *      } );
		 *    } );
		 */
		"aaData": null,
	
	
		/**
		 * If ordering is enabled, then DataTables will perform a first pass sort on
		 * initialisation. You can define which column(s) the sort is performed
		 * upon, and the sorting direction, with this variable. The `sorting` array
		 * should contain an array for each column to be sorted initially containing
		 * the column's index and a direction string ('asc' or 'desc').
		 *  @type array
		 *  @default [[0,'asc']]
		 *
		 *  @dtopt Option
		 *  @name DataTable.defaults.order
		 *
		 *  @example
		 *    // Sort by 3rd column first, and then 4th column
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "order": [[2,'asc'], [3,'desc']]
		 *      } );
		 *    } );
		 *
		 *    // No initial sorting
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "order": []
		 *      } );
		 *    } );
		 */
		"aaSorting": [[0,'asc']],
	
	
		/**
		 * This parameter is basically identical to the `sorting` parameter, but
		 * cannot be overridden by user interaction with the table. What this means
		 * is that you could have a column (visible or hidden) which the sorting
		 * will always be forced on first - any sorting after that (from the user)
		 * will then be performed as required. This can be useful for grouping rows
		 * together.
		 *  @type array
		 *  @default null
		 *
		 *  @dtopt Option
		 *  @name DataTable.defaults.orderFixed
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "orderFixed": [[0,'asc']]
		 *      } );
		 *    } )
		 */
		"aaSortingFixed": [],
	
	
		/**
		 * DataTables can be instructed to load data to display in the table from a
		 * Ajax source. This option defines how that Ajax call is made and where to.
		 *
		 * The `ajax` property has three different modes of operation, depending on
		 * how it is defined. These are:
		 *
		 * * `string` - Set the URL from where the data should be loaded from.
		 * * `object` - Define properties for `jQuery.ajax`.
		 * * `function` - Custom data get function
		 *
		 * `string`
		 * --------
		 *
		 * As a string, the `ajax` property simply defines the URL from which
		 * DataTables will load data.
		 *
		 * `object`
		 * --------
		 *
		 * As an object, the parameters in the object are passed to
		 * [jQuery.ajax](http://api.jquery.com/jQuery.ajax/) allowing fine control
		 * of the Ajax request. DataTables has a number of default parameters which
		 * you can override using this option. Please refer to the jQuery
		 * documentation for a full description of the options available, although
		 * the following parameters provide additional options in DataTables or
		 * require special consideration:
		 *
		 * * `data` - As with jQuery, `data` can be provided as an object, but it
		 *   can also be used as a function to manipulate the data DataTables sends
		 *   to the server. The function takes a single parameter, an object of
		 *   parameters with the values that DataTables has readied for sending. An
		 *   object may be returned which will be merged into the DataTables
		 *   defaults, or you can add the items to the object that was passed in and
		 *   not return anything from the function. This supersedes `fnServerParams`
		 *   from DataTables 1.9-.
		 *
		 * * `dataSrc` - By default DataTables will look for the property `data` (or
		 *   `aaData` for compatibility with DataTables 1.9-) when obtaining data
		 *   from an Ajax source or for server-side processing - this parameter
		 *   allows that property to be changed. You can use Javascript dotted
		 *   object notation to get a data source for multiple levels of nesting, or
		 *   it my be used as a function. As a function it takes a single parameter,
		 *   the JSON returned from the server, which can be manipulated as
		 *   required, with the returned value being that used by DataTables as the
		 *   data source for the table. This supersedes `sAjaxDataProp` from
		 *   DataTables 1.9-.
		 *
		 * * `success` - Should not be overridden it is used internally in
		 *   DataTables. To manipulate / transform the data returned by the server
		 *   use `ajax.dataSrc`, or use `ajax` as a function (see below).
		 *
		 * `function`
		 * ----------
		 *
		 * As a function, making the Ajax call is left up to yourself allowing
		 * complete control of the Ajax request. Indeed, if desired, a method other
		 * than Ajax could be used to obtain the required data, such as Web storage
		 * or an AIR database.
		 *
		 * The function is given four parameters and no return is required. The
		 * parameters are:
		 *
		 * 1. _object_ - Data to send to the server
		 * 2. _function_ - Callback function that must be executed when the required
		 *    data has been obtained. That data should be passed into the callback
		 *    as the only parameter
		 * 3. _object_ - DataTables settings object for the table
		 *
		 * Note that this supersedes `fnServerData` from DataTables 1.9-.
		 *
		 *  @type string|object|function
		 *  @default null
		 *
		 *  @dtopt Option
		 *  @name DataTable.defaults.ajax
		 *  @since 1.10.0
		 *
		 * @example
		 *   // Get JSON data from a file via Ajax.
		 *   // Note DataTables expects data in the form `{ data: [ ...data... ] }` by default).
		 *   $('#example').dataTable( {
		 *     "ajax": "data.json"
		 *   } );
		 *
		 * @example
		 *   // Get JSON data from a file via Ajax, using `dataSrc` to change
		 *   // `data` to `tableData` (i.e. `{ tableData: [ ...data... ] }`)
		 *   $('#example').dataTable( {
		 *     "ajax": {
		 *       "url": "data.json",
		 *       "dataSrc": "tableData"
		 *     }
		 *   } );
		 *
		 * @example
		 *   // Get JSON data from a file via Ajax, using `dataSrc` to read data
		 *   // from a plain array rather than an array in an object
		 *   $('#example').dataTable( {
		 *     "ajax": {
		 *       "url": "data.json",
		 *       "dataSrc": ""
		 *     }
		 *   } );
		 *
		 * @example
		 *   // Manipulate the data returned from the server - add a link to data
		 *   // (note this can, should, be done using `render` for the column - this
		 *   // is just a simple example of how the data can be manipulated).
		 *   $('#example').dataTable( {
		 *     "ajax": {
		 *       "url": "data.json",
		 *       "dataSrc": function ( json ) {
		 *         for ( var i=0, ien=json.length ; i<ien ; i++ ) {
		 *           json[i][0] = '<a href="/message/'+json[i][0]+'>View message</a>';
		 *         }
		 *         return json;
		 *       }
		 *     }
		 *   } );
		 *
		 * @example
		 *   // Add data to the request
		 *   $('#example').dataTable( {
		 *     "ajax": {
		 *       "url": "data.json",
		 *       "data": function ( d ) {
		 *         return {
		 *           "extra_search": $('#extra').val()
		 *         };
		 *       }
		 *     }
		 *   } );
		 *
		 * @example
		 *   // Send request as POST
		 *   $('#example').dataTable( {
		 *     "ajax": {
		 *       "url": "data.json",
		 *       "type": "POST"
		 *     }
		 *   } );
		 *
		 * @example
		 *   // Get the data from localStorage (could interface with a form for
		 *   // adding, editing and removing rows).
		 *   $('#example').dataTable( {
		 *     "ajax": function (data, callback, settings) {
		 *       callback(
		 *         JSON.parse( localStorage.getItem('dataTablesData') )
		 *       );
		 *     }
		 *   } );
		 */
		"ajax": null,
	
	
		/**
		 * This parameter allows you to readily specify the entries in the length drop
		 * down menu that DataTables shows when pagination is enabled. It can be
		 * either a 1D array of options which will be used for both the displayed
		 * option and the value, or a 2D array which will use the array in the first
		 * position as the value, and the array in the second position as the
		 * displayed options (useful for language strings such as 'All').
		 *
		 * Note that the `pageLength` property will be automatically set to the
		 * first value given in this array, unless `pageLength` is also provided.
		 *  @type array
		 *  @default [ 10, 25, 50, 100 ]
		 *
		 *  @dtopt Option
		 *  @name DataTable.defaults.lengthMenu
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "lengthMenu": [[10, 25, 50, -1], [10, 25, 50, "All"]]
		 *      } );
		 *    } );
		 */
		"aLengthMenu": [ 10, 25, 50, 100 ],
	
	
		/**
		 * The `columns` option in the initialisation parameter allows you to define
		 * details about the way individual columns behave. For a full list of
		 * column options that can be set, please see
		 * {@link DataTable.defaults.column}. Note that if you use `columns` to
		 * define your columns, you must have an entry in the array for every single
		 * column that you have in your table (these can be null if you don't which
		 * to specify any options).
		 *  @member
		 *
		 *  @name DataTable.defaults.column
		 */
		"aoColumns": null,
	
		/**
		 * Very similar to `columns`, `columnDefs` allows you to target a specific
		 * column, multiple columns, or all columns, using the `targets` property of
		 * each object in the array. This allows great flexibility when creating
		 * tables, as the `columnDefs` arrays can be of any length, targeting the
		 * columns you specifically want. `columnDefs` may use any of the column
		 * options available: {@link DataTable.defaults.column}, but it _must_
		 * have `targets` defined in each object in the array. Values in the `targets`
		 * array may be:
		 *   <ul>
		 *     <li>a string - class name will be matched on the TH for the column</li>
		 *     <li>0 or a positive integer - column index counting from the left</li>
		 *     <li>a negative integer - column index counting from the right</li>
		 *     <li>the string "_all" - all columns (i.e. assign a default)</li>
		 *   </ul>
		 *  @member
		 *
		 *  @name DataTable.defaults.columnDefs
		 */
		"aoColumnDefs": null,
	
	
		/**
		 * Basically the same as `search`, this parameter defines the individual column
		 * filtering state at initialisation time. The array must be of the same size
		 * as the number of columns, and each element be an object with the parameters
		 * `search` and `escapeRegex` (the latter is optional). 'null' is also
		 * accepted and the default will be used.
		 *  @type array
		 *  @default []
		 *
		 *  @dtopt Option
		 *  @name DataTable.defaults.searchCols
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "searchCols": [
		 *          null,
		 *          { "search": "My filter" },
		 *          null,
		 *          { "search": "^[0-9]", "escapeRegex": false }
		 *        ]
		 *      } );
		 *    } )
		 */
		"aoSearchCols": [],
	
	
		/**
		 * An array of CSS classes that should be applied to displayed rows. This
		 * array may be of any length, and DataTables will apply each class
		 * sequentially, looping when required.
		 *  @type array
		 *  @default null <i>Will take the values determined by the `oClasses.stripe*`
		 *    options</i>
		 *
		 *  @dtopt Option
		 *  @name DataTable.defaults.stripeClasses
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stripeClasses": [ 'strip1', 'strip2', 'strip3' ]
		 *      } );
		 *    } )
		 */
		"asStripeClasses": null,
	
	
		/**
		 * Enable or disable automatic column width calculation. This can be disabled
		 * as an optimisation (it takes some time to calculate the widths) if the
		 * tables widths are passed in using `columns`.
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.autoWidth
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "autoWidth": false
		 *      } );
		 *    } );
		 */
		"bAutoWidth": true,
	
	
		/**
		 * Deferred rendering can provide DataTables with a huge speed boost when you
		 * are using an Ajax or JS data source for the table. This option, when set to
		 * true, will cause DataTables to defer the creation of the table elements for
		 * each row until they are needed for a draw - saving a significant amount of
		 * time.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.deferRender
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "ajax": "sources/arrays.txt",
		 *        "deferRender": true
		 *      } );
		 *    } );
		 */
		"bDeferRender": false,
	
	
		/**
		 * Replace a DataTable which matches the given selector and replace it with
		 * one which has the properties of the new initialisation object passed. If no
		 * table matches the selector, then the new DataTable will be constructed as
		 * per normal.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.destroy
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "srollY": "200px",
		 *        "paginate": false
		 *      } );
		 *
		 *      // Some time later....
		 *      $('#example').dataTable( {
		 *        "filter": false,
		 *        "destroy": true
		 *      } );
		 *    } );
		 */
		"bDestroy": false,
	
	
		/**
		 * Enable or disable filtering of data. Filtering in DataTables is "smart" in
		 * that it allows the end user to input multiple words (space separated) and
		 * will match a row containing those words, even if not in the order that was
		 * specified (this allow matching across multiple columns). Note that if you
		 * wish to use filtering in DataTables this must remain 'true' - to remove the
		 * default filtering input box and retain filtering abilities, please use
		 * {@link DataTable.defaults.dom}.
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.searching
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "searching": false
		 *      } );
		 *    } );
		 */
		"bFilter": true,
	
	
		/**
		 * Enable or disable the table information display. This shows information
		 * about the data that is currently visible on the page, including information
		 * about filtered data if that action is being performed.
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.info
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "info": false
		 *      } );
		 *    } );
		 */
		"bInfo": true,
	
	
		/**
		 * Allows the end user to select the size of a formatted page from a select
		 * menu (sizes are 10, 25, 50 and 100). Requires pagination (`paginate`).
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.lengthChange
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "lengthChange": false
		 *      } );
		 *    } );
		 */
		"bLengthChange": true,
	
	
		/**
		 * Enable or disable pagination.
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.paging
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "paging": false
		 *      } );
		 *    } );
		 */
		"bPaginate": true,
	
	
		/**
		 * Enable or disable the display of a 'processing' indicator when the table is
		 * being processed (e.g. a sort). This is particularly useful for tables with
		 * large amounts of data where it can take a noticeable amount of time to sort
		 * the entries.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.processing
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "processing": true
		 *      } );
		 *    } );
		 */
		"bProcessing": false,
	
	
		/**
		 * Retrieve the DataTables object for the given selector. Note that if the
		 * table has already been initialised, this parameter will cause DataTables
		 * to simply return the object that has already been set up - it will not take
		 * account of any changes you might have made to the initialisation object
		 * passed to DataTables (setting this parameter to true is an acknowledgement
		 * that you understand this). `destroy` can be used to reinitialise a table if
		 * you need.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.retrieve
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      initTable();
		 *      tableActions();
		 *    } );
		 *
		 *    function initTable ()
		 *    {
		 *      return $('#example').dataTable( {
		 *        "scrollY": "200px",
		 *        "paginate": false,
		 *        "retrieve": true
		 *      } );
		 *    }
		 *
		 *    function tableActions ()
		 *    {
		 *      var table = initTable();
		 *      // perform API operations with oTable
		 *    }
		 */
		"bRetrieve": false,
	
	
		/**
		 * When vertical (y) scrolling is enabled, DataTables will force the height of
		 * the table's viewport to the given height at all times (useful for layout).
		 * However, this can look odd when filtering data down to a small data set,
		 * and the footer is left "floating" further down. This parameter (when
		 * enabled) will cause DataTables to collapse the table's viewport down when
		 * the result set will fit within the given Y height.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.scrollCollapse
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "scrollY": "200",
		 *        "scrollCollapse": true
		 *      } );
		 *    } );
		 */
		"bScrollCollapse": false,
	
	
		/**
		 * Configure DataTables to use server-side processing. Note that the
		 * `ajax` parameter must also be given in order to give DataTables a
		 * source to obtain the required data for each draw.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Features
		 *  @dtopt Server-side
		 *  @name DataTable.defaults.serverSide
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "serverSide": true,
		 *        "ajax": "xhr.php"
		 *      } );
		 *    } );
		 */
		"bServerSide": false,
	
	
		/**
		 * Enable or disable sorting of columns. Sorting of individual columns can be
		 * disabled by the `sortable` option for each column.
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.ordering
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "ordering": false
		 *      } );
		 *    } );
		 */
		"bSort": true,
	
	
		/**
		 * Enable or display DataTables' ability to sort multiple columns at the
		 * same time (activated by shift-click by the user).
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.orderMulti
		 *
		 *  @example
		 *    // Disable multiple column sorting ability
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "orderMulti": false
		 *      } );
		 *    } );
		 */
		"bSortMulti": true,
	
	
		/**
		 * Allows control over whether DataTables should use the top (true) unique
		 * cell that is found for a single column, or the bottom (false - default).
		 * This is useful when using complex headers.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.orderCellsTop
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "orderCellsTop": true
		 *      } );
		 *    } );
		 */
		"bSortCellsTop": false,
	
	
		/**
		 * Enable or disable the addition of the classes `sorting\_1`, `sorting\_2` and
		 * `sorting\_3` to the columns which are currently being sorted on. This is
		 * presented as a feature switch as it can increase processing time (while
		 * classes are removed and added) so for large data sets you might want to
		 * turn this off.
		 *  @type boolean
		 *  @default true
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.orderClasses
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "orderClasses": false
		 *      } );
		 *    } );
		 */
		"bSortClasses": true,
	
	
		/**
		 * Enable or disable state saving. When enabled HTML5 `localStorage` will be
		 * used to save table display information such as pagination information,
		 * display length, filtering and sorting. As such when the end user reloads
		 * the page the display display will match what thy had previously set up.
		 *
		 * Due to the use of `localStorage` the default state saving is not supported
		 * in IE6 or 7. If state saving is required in those browsers, use
		 * `stateSaveCallback` to provide a storage solution such as cookies.
		 *  @type boolean
		 *  @default false
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.stateSave
		 *
		 *  @example
		 *    $(document).ready( function () {
		 *      $('#example').dataTable( {
		 *        "stateSave": true
		 *      } );
		 *    } );
		 */
		"bStateSave": false,
	
	
		/**
		 * This function is called when a TR element is created (and all TD child
		 * elements have been inserted), or registered if using a DOM source, allowing
		 * manipulation of the TR element (adding classes etc).
		 *  @type function
		 *  @param {node} row "TR" element for the current row
		 *  @param {array} data Raw data array for this row
		 *  @param {int} dataIndex The index of this row in the internal aoData array
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.createdRow
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "createdRow": function( row, data, dataIndex ) {
		 *          // Bold the grade for all 'A' grade browsers
		 *          if ( data[4] == "A" )
		 *          {
		 *            $('td:eq(4)', row).html( '<b>A</b>' );
		 *          }
		 *        }
		 *      } );
		 *    } );
		 */
		"fnCreatedRow": null,
	
	
		/**
		 * This function is called on every 'draw' event, and allows you to
		 * dynamically modify any aspect you want about the created DOM.
		 *  @type function
		 *  @param {object} settings DataTables settings object
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.drawCallback
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "drawCallback": function( settings ) {
		 *          alert( 'DataTables has redrawn the table' );
		 *        }
		 *      } );
		 *    } );
		 */
		"fnDrawCallback": null,
	
	
		/**
		 * Identical to fnHeaderCallback() but for the table footer this function
		 * allows you to modify the table footer on every 'draw' event.
		 *  @type function
		 *  @param {node} foot "TR" element for the footer
		 *  @param {array} data Full table data (as derived from the original HTML)
		 *  @param {int} start Index for the current display starting point in the
		 *    display array
		 *  @param {int} end Index for the current display ending point in the
		 *    display array
		 *  @param {array int} display Index array to translate the visual position
		 *    to the full data array
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.footerCallback
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "footerCallback": function( tfoot, data, start, end, display ) {
		 *          tfoot.getElementsByTagName('th')[0].innerHTML = "Starting index is "+start;
		 *        }
		 *      } );
		 *    } )
		 */
		"fnFooterCallback": null,
	
	
		/**
		 * When rendering large numbers in the information element for the table
		 * (i.e. "Showing 1 to 10 of 57 entries") DataTables will render large numbers
		 * to have a comma separator for the 'thousands' units (e.g. 1 million is
		 * rendered as "1,000,000") to help readability for the end user. This
		 * function will override the default method DataTables uses.
		 *  @type function
		 *  @member
		 *  @param {int} toFormat number to be formatted
		 *  @returns {string} formatted string for DataTables to show the number
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.formatNumber
		 *
		 *  @example
		 *    // Format a number using a single quote for the separator (note that
		 *    // this can also be done with the language.thousands option)
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "formatNumber": function ( toFormat ) {
		 *          return toFormat.toString().replace(
		 *            /\B(?=(\d{3})+(?!\d))/g, "'"
		 *          );
		 *        };
		 *      } );
		 *    } );
		 */
		"fnFormatNumber": function ( toFormat ) {
			return toFormat.toString().replace(
				/\B(?=(\d{3})+(?!\d))/g,
				this.oLanguage.sThousands
			);
		},
	
	
		/**
		 * This function is called on every 'draw' event, and allows you to
		 * dynamically modify the header row. This can be used to calculate and
		 * display useful information about the table.
		 *  @type function
		 *  @param {node} head "TR" element for the header
		 *  @param {array} data Full table data (as derived from the original HTML)
		 *  @param {int} start Index for the current display starting point in the
		 *    display array
		 *  @param {int} end Index for the current display ending point in the
		 *    display array
		 *  @param {array int} display Index array to translate the visual position
		 *    to the full data array
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.headerCallback
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "fheaderCallback": function( head, data, start, end, display ) {
		 *          head.getElementsByTagName('th')[0].innerHTML = "Displaying "+(end-start)+" records";
		 *        }
		 *      } );
		 *    } )
		 */
		"fnHeaderCallback": null,
	
	
		/**
		 * The information element can be used to convey information about the current
		 * state of the table. Although the internationalisation options presented by
		 * DataTables are quite capable of dealing with most customisations, there may
		 * be times where you wish to customise the string further. This callback
		 * allows you to do exactly that.
		 *  @type function
		 *  @param {object} oSettings DataTables settings object
		 *  @param {int} start Starting position in data for the draw
		 *  @param {int} end End position in data for the draw
		 *  @param {int} max Total number of rows in the table (regardless of
		 *    filtering)
		 *  @param {int} total Total number of rows in the data set, after filtering
		 *  @param {string} pre The string that DataTables has formatted using it's
		 *    own rules
		 *  @returns {string} The string to be displayed in the information element.
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.infoCallback
		 *
		 *  @example
		 *    $('#example').dataTable( {
		 *      "infoCallback": function( settings, start, end, max, total, pre ) {
		 *        return start +" to "+ end;
		 *      }
		 *    } );
		 */
		"fnInfoCallback": null,
	
	
		/**
		 * Called when the table has been initialised. Normally DataTables will
		 * initialise sequentially and there will be no need for this function,
		 * however, this does not hold true when using external language information
		 * since that is obtained using an async XHR call.
		 *  @type function
		 *  @param {object} settings DataTables settings object
		 *  @param {object} json The JSON object request from the server - only
		 *    present if client-side Ajax sourced data is used
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.initComplete
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "initComplete": function(settings, json) {
		 *          alert( 'DataTables has finished its initialisation.' );
		 *        }
		 *      } );
		 *    } )
		 */
		"fnInitComplete": null,
	
	
		/**
		 * Called at the very start of each table draw and can be used to cancel the
		 * draw by returning false, any other return (including undefined) results in
		 * the full draw occurring).
		 *  @type function
		 *  @param {object} settings DataTables settings object
		 *  @returns {boolean} False will cancel the draw, anything else (including no
		 *    return) will allow it to complete.
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.preDrawCallback
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "preDrawCallback": function( settings ) {
		 *          if ( $('#test').val() == 1 ) {
		 *            return false;
		 *          }
		 *        }
		 *      } );
		 *    } );
		 */
		"fnPreDrawCallback": null,
	
	
		/**
		 * This function allows you to 'post process' each row after it have been
		 * generated for each table draw, but before it is rendered on screen. This
		 * function might be used for setting the row class name etc.
		 *  @type function
		 *  @param {node} row "TR" element for the current row
		 *  @param {array} data Raw data array for this row
		 *  @param {int} displayIndex The display index for the current table draw
		 *  @param {int} displayIndexFull The index of the data in the full list of
		 *    rows (after filtering)
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.rowCallback
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "rowCallback": function( row, data, displayIndex, displayIndexFull ) {
		 *          // Bold the grade for all 'A' grade browsers
		 *          if ( data[4] == "A" ) {
		 *            $('td:eq(4)', row).html( '<b>A</b>' );
		 *          }
		 *        }
		 *      } );
		 *    } );
		 */
		"fnRowCallback": null,
	
	
		/**
		 * __Deprecated__ The functionality provided by this parameter has now been
		 * superseded by that provided through `ajax`, which should be used instead.
		 *
		 * This parameter allows you to override the default function which obtains
		 * the data from the server so something more suitable for your application.
		 * For example you could use POST data, or pull information from a Gears or
		 * AIR database.
		 *  @type function
		 *  @member
		 *  @param {string} source HTTP source to obtain the data from (`ajax`)
		 *  @param {array} data A key/value pair object containing the data to send
		 *    to the server
		 *  @param {function} callback to be called on completion of the data get
		 *    process that will draw the data on the page.
		 *  @param {object} settings DataTables settings object
		 *
		 *  @dtopt Callbacks
		 *  @dtopt Server-side
		 *  @name DataTable.defaults.serverData
		 *
		 *  @deprecated 1.10. Please use `ajax` for this functionality now.
		 */
		"fnServerData": null,
	
	
		/**
		 * __Deprecated__ The functionality provided by this parameter has now been
		 * superseded by that provided through `ajax`, which should be used instead.
		 *
		 *  It is often useful to send extra data to the server when making an Ajax
		 * request - for example custom filtering information, and this callback
		 * function makes it trivial to send extra information to the server. The
		 * passed in parameter is the data set that has been constructed by
		 * DataTables, and you can add to this or modify it as you require.
		 *  @type function
		 *  @param {array} data Data array (array of objects which are name/value
		 *    pairs) that has been constructed by DataTables and will be sent to the
		 *    server. In the case of Ajax sourced data with server-side processing
		 *    this will be an empty array, for server-side processing there will be a
		 *    significant number of parameters!
		 *  @returns {undefined} Ensure that you modify the data array passed in,
		 *    as this is passed by reference.
		 *
		 *  @dtopt Callbacks
		 *  @dtopt Server-side
		 *  @name DataTable.defaults.serverParams
		 *
		 *  @deprecated 1.10. Please use `ajax` for this functionality now.
		 */
		"fnServerParams": null,
	
	
		/**
		 * Load the table state. With this function you can define from where, and how, the
		 * state of a table is loaded. By default DataTables will load from `localStorage`
		 * but you might wish to use a server-side database or cookies.
		 *  @type function
		 *  @member
		 *  @param {object} settings DataTables settings object
		 *  @param {object} callback Callback that can be executed when done. It
		 *    should be passed the loaded state object.
		 *  @return {object} The DataTables state object to be loaded
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.stateLoadCallback
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stateSave": true,
		 *        "stateLoadCallback": function (settings, callback) {
		 *          $.ajax( {
		 *            "url": "/state_load",
		 *            "dataType": "json",
		 *            "success": function (json) {
		 *              callback( json );
		 *            }
		 *          } );
		 *        }
		 *      } );
		 *    } );
		 */
		"fnStateLoadCallback": function ( settings ) {
			try {
				return JSON.parse(
					(settings.iStateDuration === -1 ? sessionStorage : localStorage).getItem(
						'DataTables_'+settings.sInstance+'_'+location.pathname
					)
				);
			} catch (e) {
				return {};
			}
		},
	
	
		/**
		 * Callback which allows modification of the saved state prior to loading that state.
		 * This callback is called when the table is loading state from the stored data, but
		 * prior to the settings object being modified by the saved state. Note that for
		 * plug-in authors, you should use the `stateLoadParams` event to load parameters for
		 * a plug-in.
		 *  @type function
		 *  @param {object} settings DataTables settings object
		 *  @param {object} data The state object that is to be loaded
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.stateLoadParams
		 *
		 *  @example
		 *    // Remove a saved filter, so filtering is never loaded
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stateSave": true,
		 *        "stateLoadParams": function (settings, data) {
		 *          data.oSearch.sSearch = "";
		 *        }
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Disallow state loading by returning false
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stateSave": true,
		 *        "stateLoadParams": function (settings, data) {
		 *          return false;
		 *        }
		 *      } );
		 *    } );
		 */
		"fnStateLoadParams": null,
	
	
		/**
		 * Callback that is called when the state has been loaded from the state saving method
		 * and the DataTables settings object has been modified as a result of the loaded state.
		 *  @type function
		 *  @param {object} settings DataTables settings object
		 *  @param {object} data The state object that was loaded
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.stateLoaded
		 *
		 *  @example
		 *    // Show an alert with the filtering value that was saved
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stateSave": true,
		 *        "stateLoaded": function (settings, data) {
		 *          alert( 'Saved filter was: '+data.oSearch.sSearch );
		 *        }
		 *      } );
		 *    } );
		 */
		"fnStateLoaded": null,
	
	
		/**
		 * Save the table state. This function allows you to define where and how the state
		 * information for the table is stored By default DataTables will use `localStorage`
		 * but you might wish to use a server-side database or cookies.
		 *  @type function
		 *  @member
		 *  @param {object} settings DataTables settings object
		 *  @param {object} data The state object to be saved
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.stateSaveCallback
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stateSave": true,
		 *        "stateSaveCallback": function (settings, data) {
		 *          // Send an Ajax request to the server with the state object
		 *          $.ajax( {
		 *            "url": "/state_save",
		 *            "data": data,
		 *            "dataType": "json",
		 *            "method": "POST"
		 *            "success": function () {}
		 *          } );
		 *        }
		 *      } );
		 *    } );
		 */
		"fnStateSaveCallback": function ( settings, data ) {
			try {
				(settings.iStateDuration === -1 ? sessionStorage : localStorage).setItem(
					'DataTables_'+settings.sInstance+'_'+location.pathname,
					JSON.stringify( data )
				);
			} catch (e) {}
		},
	
	
		/**
		 * Callback which allows modification of the state to be saved. Called when the table
		 * has changed state a new state save is required. This method allows modification of
		 * the state saving object prior to actually doing the save, including addition or
		 * other state properties or modification. Note that for plug-in authors, you should
		 * use the `stateSaveParams` event to save parameters for a plug-in.
		 *  @type function
		 *  @param {object} settings DataTables settings object
		 *  @param {object} data The state object to be saved
		 *
		 *  @dtopt Callbacks
		 *  @name DataTable.defaults.stateSaveParams
		 *
		 *  @example
		 *    // Remove a saved filter, so filtering is never saved
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stateSave": true,
		 *        "stateSaveParams": function (settings, data) {
		 *          data.oSearch.sSearch = "";
		 *        }
		 *      } );
		 *    } );
		 */
		"fnStateSaveParams": null,
	
	
		/**
		 * Duration for which the saved state information is considered valid. After this period
		 * has elapsed the state will be returned to the default.
		 * Value is given in seconds.
		 *  @type int
		 *  @default 7200 <i>(2 hours)</i>
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.stateDuration
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "stateDuration": 60*60*24; // 1 day
		 *      } );
		 *    } )
		 */
		"iStateDuration": 7200,
	
	
		/**
		 * When enabled DataTables will not make a request to the server for the first
		 * page draw - rather it will use the data already on the page (no sorting etc
		 * will be applied to it), thus saving on an XHR at load time. `deferLoading`
		 * is used to indicate that deferred loading is required, but it is also used
		 * to tell DataTables how many records there are in the full table (allowing
		 * the information element and pagination to be displayed correctly). In the case
		 * where a filtering is applied to the table on initial load, this can be
		 * indicated by giving the parameter as an array, where the first element is
		 * the number of records available after filtering and the second element is the
		 * number of records without filtering (allowing the table information element
		 * to be shown correctly).
		 *  @type int | array
		 *  @default null
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.deferLoading
		 *
		 *  @example
		 *    // 57 records available in the table, no filtering applied
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "serverSide": true,
		 *        "ajax": "scripts/server_processing.php",
		 *        "deferLoading": 57
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // 57 records after filtering, 100 without filtering (an initial filter applied)
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "serverSide": true,
		 *        "ajax": "scripts/server_processing.php",
		 *        "deferLoading": [ 57, 100 ],
		 *        "search": {
		 *          "search": "my_filter"
		 *        }
		 *      } );
		 *    } );
		 */
		"iDeferLoading": null,
	
	
		/**
		 * Number of rows to display on a single page when using pagination. If
		 * feature enabled (`lengthChange`) then the end user will be able to override
		 * this to a custom setting using a pop-up menu.
		 *  @type int
		 *  @default 10
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.pageLength
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "pageLength": 50
		 *      } );
		 *    } )
		 */
		"iDisplayLength": 10,
	
	
		/**
		 * Define the starting point for data display when using DataTables with
		 * pagination. Note that this parameter is the number of records, rather than
		 * the page number, so if you have 10 records per page and want to start on
		 * the third page, it should be "20".
		 *  @type int
		 *  @default 0
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.displayStart
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "displayStart": 20
		 *      } );
		 *    } )
		 */
		"iDisplayStart": 0,
	
	
		/**
		 * By default DataTables allows keyboard navigation of the table (sorting, paging,
		 * and filtering) by adding a `tabindex` attribute to the required elements. This
		 * allows you to tab through the controls and press the enter key to activate them.
		 * The tabindex is default 0, meaning that the tab follows the flow of the document.
		 * You can overrule this using this parameter if you wish. Use a value of -1 to
		 * disable built-in keyboard navigation.
		 *  @type int
		 *  @default 0
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.tabIndex
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "tabIndex": 1
		 *      } );
		 *    } );
		 */
		"iTabIndex": 0,
	
	
		/**
		 * Classes that DataTables assigns to the various components and features
		 * that it adds to the HTML table. This allows classes to be configured
		 * during initialisation in addition to through the static
		 * {@link DataTable.ext.oStdClasses} object).
		 *  @namespace
		 *  @name DataTable.defaults.classes
		 */
		"oClasses": {},
	
	
		/**
		 * All strings that DataTables uses in the user interface that it creates
		 * are defined in this object, allowing you to modified them individually or
		 * completely replace them all as required.
		 *  @namespace
		 *  @name DataTable.defaults.language
		 */
		"oLanguage": {
			/**
			 * Strings that are used for WAI-ARIA labels and controls only (these are not
			 * actually visible on the page, but will be read by screenreaders, and thus
			 * must be internationalised as well).
			 *  @namespace
			 *  @name DataTable.defaults.language.aria
			 */
			"oAria": {
				/**
				 * ARIA label that is added to the table headers when the column may be
				 * sorted ascending by activing the column (click or return when focused).
				 * Note that the column header is prefixed to this string.
				 *  @type string
				 *  @default : activate to sort column ascending
				 *
				 *  @dtopt Language
				 *  @name DataTable.defaults.language.aria.sortAscending
				 *
				 *  @example
				 *    $(document).ready( function() {
				 *      $('#example').dataTable( {
				 *        "language": {
				 *          "aria": {
				 *            "sortAscending": " - click/return to sort ascending"
				 *          }
				 *        }
				 *      } );
				 *    } );
				 */
				"sSortAscending": ": activate to sort column ascending",
	
				/**
				 * ARIA label that is added to the table headers when the column may be
				 * sorted descending by activing the column (click or return when focused).
				 * Note that the column header is prefixed to this string.
				 *  @type string
				 *  @default : activate to sort column ascending
				 *
				 *  @dtopt Language
				 *  @name DataTable.defaults.language.aria.sortDescending
				 *
				 *  @example
				 *    $(document).ready( function() {
				 *      $('#example').dataTable( {
				 *        "language": {
				 *          "aria": {
				 *            "sortDescending": " - click/return to sort descending"
				 *          }
				 *        }
				 *      } );
				 *    } );
				 */
				"sSortDescending": ": activate to sort column descending"
			},
	
			/**
			 * Pagination string used by DataTables for the built-in pagination
			 * control types.
			 *  @namespace
			 *  @name DataTable.defaults.language.paginate
			 */
			"oPaginate": {
				/**
				 * Text to use when using the 'full_numbers' type of pagination for the
				 * button to take the user to the first page.
				 *  @type string
				 *  @default First
				 *
				 *  @dtopt Language
				 *  @name DataTable.defaults.language.paginate.first
				 *
				 *  @example
				 *    $(document).ready( function() {
				 *      $('#example').dataTable( {
				 *        "language": {
				 *          "paginate": {
				 *            "first": "First page"
				 *          }
				 *        }
				 *      } );
				 *    } );
				 */
				"sFirst": "First",
	
	
				/**
				 * Text to use when using the 'full_numbers' type of pagination for the
				 * button to take the user to the last page.
				 *  @type string
				 *  @default Last
				 *
				 *  @dtopt Language
				 *  @name DataTable.defaults.language.paginate.last
				 *
				 *  @example
				 *    $(document).ready( function() {
				 *      $('#example').dataTable( {
				 *        "language": {
				 *          "paginate": {
				 *            "last": "Last page"
				 *          }
				 *        }
				 *      } );
				 *    } );
				 */
				"sLast": "Last",
	
	
				/**
				 * Text to use for the 'next' pagination button (to take the user to the
				 * next page).
				 *  @type string
				 *  @default Next
				 *
				 *  @dtopt Language
				 *  @name DataTable.defaults.language.paginate.next
				 *
				 *  @example
				 *    $(document).ready( function() {
				 *      $('#example').dataTable( {
				 *        "language": {
				 *          "paginate": {
				 *            "next": "Next page"
				 *          }
				 *        }
				 *      } );
				 *    } );
				 */
				"sNext": "Next",
	
	
				/**
				 * Text to use for the 'previous' pagination button (to take the user to
				 * the previous page).
				 *  @type string
				 *  @default Previous
				 *
				 *  @dtopt Language
				 *  @name DataTable.defaults.language.paginate.previous
				 *
				 *  @example
				 *    $(document).ready( function() {
				 *      $('#example').dataTable( {
				 *        "language": {
				 *          "paginate": {
				 *            "previous": "Previous page"
				 *          }
				 *        }
				 *      } );
				 *    } );
				 */
				"sPrevious": "Previous"
			},
	
			/**
			 * This string is shown in preference to `zeroRecords` when the table is
			 * empty of data (regardless of filtering). Note that this is an optional
			 * parameter - if it is not given, the value of `zeroRecords` will be used
			 * instead (either the default or given value).
			 *  @type string
			 *  @default No data available in table
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.emptyTable
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "emptyTable": "No data available in table"
			 *        }
			 *      } );
			 *    } );
			 */
			"sEmptyTable": "No data available in table",
	
	
			/**
			 * This string gives information to the end user about the information
			 * that is current on display on the page. The following tokens can be
			 * used in the string and will be dynamically replaced as the table
			 * display updates. This tokens can be placed anywhere in the string, or
			 * removed as needed by the language requires:
			 *
			 * * `\_START\_` - Display index of the first record on the current page
			 * * `\_END\_` - Display index of the last record on the current page
			 * * `\_TOTAL\_` - Number of records in the table after filtering
			 * * `\_MAX\_` - Number of records in the table without filtering
			 * * `\_PAGE\_` - Current page number
			 * * `\_PAGES\_` - Total number of pages of data in the table
			 *
			 *  @type string
			 *  @default Showing _START_ to _END_ of _TOTAL_ entries
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.info
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "info": "Showing page _PAGE_ of _PAGES_"
			 *        }
			 *      } );
			 *    } );
			 */
			"sInfo": "Showing _START_ to _END_ of _TOTAL_ entries",
	
	
			/**
			 * Display information string for when the table is empty. Typically the
			 * format of this string should match `info`.
			 *  @type string
			 *  @default Showing 0 to 0 of 0 entries
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.infoEmpty
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "infoEmpty": "No entries to show"
			 *        }
			 *      } );
			 *    } );
			 */
			"sInfoEmpty": "Showing 0 to 0 of 0 entries",
	
	
			/**
			 * When a user filters the information in a table, this string is appended
			 * to the information (`info`) to give an idea of how strong the filtering
			 * is. The variable _MAX_ is dynamically updated.
			 *  @type string
			 *  @default (filtered from _MAX_ total entries)
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.infoFiltered
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "infoFiltered": " - filtering from _MAX_ records"
			 *        }
			 *      } );
			 *    } );
			 */
			"sInfoFiltered": "(filtered from _MAX_ total entries)",
	
	
			/**
			 * If can be useful to append extra information to the info string at times,
			 * and this variable does exactly that. This information will be appended to
			 * the `info` (`infoEmpty` and `infoFiltered` in whatever combination they are
			 * being used) at all times.
			 *  @type string
			 *  @default <i>Empty string</i>
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.infoPostFix
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "infoPostFix": "All records shown are derived from real information."
			 *        }
			 *      } );
			 *    } );
			 */
			"sInfoPostFix": "",
	
	
			/**
			 * This decimal place operator is a little different from the other
			 * language options since DataTables doesn't output floating point
			 * numbers, so it won't ever use this for display of a number. Rather,
			 * what this parameter does is modify the sort methods of the table so
			 * that numbers which are in a format which has a character other than
			 * a period (`.`) as a decimal place will be sorted numerically.
			 *
			 * Note that numbers with different decimal places cannot be shown in
			 * the same table and still be sortable, the table must be consistent.
			 * However, multiple different tables on the page can use different
			 * decimal place characters.
			 *  @type string
			 *  @default 
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.decimal
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "decimal": ","
			 *          "thousands": "."
			 *        }
			 *      } );
			 *    } );
			 */
			"sDecimal": "",
	
	
			/**
			 * DataTables has a build in number formatter (`formatNumber`) which is
			 * used to format large numbers that are used in the table information.
			 * By default a comma is used, but this can be trivially changed to any
			 * character you wish with this parameter.
			 *  @type string
			 *  @default ,
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.thousands
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "thousands": "'"
			 *        }
			 *      } );
			 *    } );
			 */
			"sThousands": ",",
	
	
			/**
			 * Detail the action that will be taken when the drop down menu for the
			 * pagination length option is changed. The '_MENU_' variable is replaced
			 * with a default select list of 10, 25, 50 and 100, and can be replaced
			 * with a custom select box if required.
			 *  @type string
			 *  @default Show _MENU_ entries
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.lengthMenu
			 *
			 *  @example
			 *    // Language change only
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "lengthMenu": "Display _MENU_ records"
			 *        }
			 *      } );
			 *    } );
			 *
			 *  @example
			 *    // Language and options change
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "lengthMenu": 'Display <select>'+
			 *            '<option value="10">10</option>'+
			 *            '<option value="20">20</option>'+
			 *            '<option value="30">30</option>'+
			 *            '<option value="40">40</option>'+
			 *            '<option value="50">50</option>'+
			 *            '<option value="-1">All</option>'+
			 *            '</select> records'
			 *        }
			 *      } );
			 *    } );
			 */
			"sLengthMenu": "Show _MENU_ entries",
	
	
			/**
			 * When using Ajax sourced data and during the first draw when DataTables is
			 * gathering the data, this message is shown in an empty row in the table to
			 * indicate to the end user the the data is being loaded. Note that this
			 * parameter is not used when loading data by server-side processing, just
			 * Ajax sourced data with client-side processing.
			 *  @type string
			 *  @default Loading...
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.loadingRecords
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "loadingRecords": "Please wait - loading..."
			 *        }
			 *      } );
			 *    } );
			 */
			"sLoadingRecords": "Loading...",
	
	
			/**
			 * Text which is displayed when the table is processing a user action
			 * (usually a sort command or similar).
			 *  @type string
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.processing
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "processing": "DataTables is currently busy"
			 *        }
			 *      } );
			 *    } );
			 */
			"sProcessing": "",
	
	
			/**
			 * Details the actions that will be taken when the user types into the
			 * filtering input text box. The variable "_INPUT_", if used in the string,
			 * is replaced with the HTML text box for the filtering input allowing
			 * control over where it appears in the string. If "_INPUT_" is not given
			 * then the input box is appended to the string automatically.
			 *  @type string
			 *  @default Search:
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.search
			 *
			 *  @example
			 *    // Input text box will be appended at the end automatically
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "search": "Filter records:"
			 *        }
			 *      } );
			 *    } );
			 *
			 *  @example
			 *    // Specify where the filter should appear
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "search": "Apply filter _INPUT_ to table"
			 *        }
			 *      } );
			 *    } );
			 */
			"sSearch": "Search:",
	
	
			/**
			 * Assign a `placeholder` attribute to the search `input` element
			 *  @type string
			 *  @default 
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.searchPlaceholder
			 */
			"sSearchPlaceholder": "",
	
	
			/**
			 * All of the language information can be stored in a file on the
			 * server-side, which DataTables will look up if this parameter is passed.
			 * It must store the URL of the language file, which is in a JSON format,
			 * and the object has the same properties as the oLanguage object in the
			 * initialiser object (i.e. the above parameters). Please refer to one of
			 * the example language files to see how this works in action.
			 *  @type string
			 *  @default <i>Empty string - i.e. disabled</i>
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.url
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "url": "http://www.sprymedia.co.uk/dataTables/lang.txt"
			 *        }
			 *      } );
			 *    } );
			 */
			"sUrl": "",
	
	
			/**
			 * Text shown inside the table records when the is no information to be
			 * displayed after filtering. `emptyTable` is shown when there is simply no
			 * information in the table at all (regardless of filtering).
			 *  @type string
			 *  @default No matching records found
			 *
			 *  @dtopt Language
			 *  @name DataTable.defaults.language.zeroRecords
			 *
			 *  @example
			 *    $(document).ready( function() {
			 *      $('#example').dataTable( {
			 *        "language": {
			 *          "zeroRecords": "No records to display"
			 *        }
			 *      } );
			 *    } );
			 */
			"sZeroRecords": "No matching records found"
		},
	
	
		/**
		 * This parameter allows you to have define the global filtering state at
		 * initialisation time. As an object the `search` parameter must be
		 * defined, but all other parameters are optional. When `regex` is true,
		 * the search string will be treated as a regular expression, when false
		 * (default) it will be treated as a straight string. When `smart`
		 * DataTables will use it's smart filtering methods (to word match at
		 * any point in the data), when false this will not be done.
		 *  @namespace
		 *  @extends DataTable.models.oSearch
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.search
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "search": {"search": "Initial search"}
		 *      } );
		 *    } )
		 */
		"oSearch": $.extend( {}, DataTable.models.oSearch ),
	
	
		/**
		 * __Deprecated__ The functionality provided by this parameter has now been
		 * superseded by that provided through `ajax`, which should be used instead.
		 *
		 * By default DataTables will look for the property `data` (or `aaData` for
		 * compatibility with DataTables 1.9-) when obtaining data from an Ajax
		 * source or for server-side processing - this parameter allows that
		 * property to be changed. You can use Javascript dotted object notation to
		 * get a data source for multiple levels of nesting.
		 *  @type string
		 *  @default data
		 *
		 *  @dtopt Options
		 *  @dtopt Server-side
		 *  @name DataTable.defaults.ajaxDataProp
		 *
		 *  @deprecated 1.10. Please use `ajax` for this functionality now.
		 */
		"sAjaxDataProp": "data",
	
	
		/**
		 * __Deprecated__ The functionality provided by this parameter has now been
		 * superseded by that provided through `ajax`, which should be used instead.
		 *
		 * You can instruct DataTables to load data from an external
		 * source using this parameter (use aData if you want to pass data in you
		 * already have). Simply provide a url a JSON object can be obtained from.
		 *  @type string
		 *  @default null
		 *
		 *  @dtopt Options
		 *  @dtopt Server-side
		 *  @name DataTable.defaults.ajaxSource
		 *
		 *  @deprecated 1.10. Please use `ajax` for this functionality now.
		 */
		"sAjaxSource": null,
	
	
		/**
		 * This initialisation variable allows you to specify exactly where in the
		 * DOM you want DataTables to inject the various controls it adds to the page
		 * (for example you might want the pagination controls at the top of the
		 * table). DIV elements (with or without a custom class) can also be added to
		 * aid styling. The follow syntax is used:
		 *   <ul>
		 *     <li>The following options are allowed:
		 *       <ul>
		 *         <li>'l' - Length changing</li>
		 *         <li>'f' - Filtering input</li>
		 *         <li>'t' - The table!</li>
		 *         <li>'i' - Information</li>
		 *         <li>'p' - Pagination</li>
		 *         <li>'r' - pRocessing</li>
		 *       </ul>
		 *     </li>
		 *     <li>The following constants are allowed:
		 *       <ul>
		 *         <li>'H' - jQueryUI theme "header" classes ('fg-toolbar ui-widget-header ui-corner-tl ui-corner-tr ui-helper-clearfix')</li>
		 *         <li>'F' - jQueryUI theme "footer" classes ('fg-toolbar ui-widget-header ui-corner-bl ui-corner-br ui-helper-clearfix')</li>
		 *       </ul>
		 *     </li>
		 *     <li>The following syntax is expected:
		 *       <ul>
		 *         <li>'&lt;' and '&gt;' - div elements</li>
		 *         <li>'&lt;"class" and '&gt;' - div with a class</li>
		 *         <li>'&lt;"#id" and '&gt;' - div with an ID</li>
		 *       </ul>
		 *     </li>
		 *     <li>Examples:
		 *       <ul>
		 *         <li>'&lt;"wrapper"flipt&gt;'</li>
		 *         <li>'&lt;lf&lt;t&gt;ip&gt;'</li>
		 *       </ul>
		 *     </li>
		 *   </ul>
		 *  @type string
		 *  @default lfrtip <i>(when `jQueryUI` is false)</i> <b>or</b>
		 *    <"H"lfr>t<"F"ip> <i>(when `jQueryUI` is true)</i>
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.dom
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "dom": '&lt;"top"i&gt;rt&lt;"bottom"flp&gt;&lt;"clear"&gt;'
		 *      } );
		 *    } );
		 */
		"sDom": "lfrtip",
	
	
		/**
		 * Search delay option. This will throttle full table searches that use the
		 * DataTables provided search input element (it does not effect calls to
		 * `dt-api search()`, providing a delay before the search is made.
		 *  @type integer
		 *  @default 0
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.searchDelay
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "searchDelay": 200
		 *      } );
		 *    } )
		 */
		"searchDelay": null,
	
	
		/**
		 * DataTables features six different built-in options for the buttons to
		 * display for pagination control:
		 *
		 * * `numbers` - Page number buttons only
		 * * `simple` - 'Previous' and 'Next' buttons only
		 * * 'simple_numbers` - 'Previous' and 'Next' buttons, plus page numbers
		 * * `full` - 'First', 'Previous', 'Next' and 'Last' buttons
		 * * `full_numbers` - 'First', 'Previous', 'Next' and 'Last' buttons, plus page numbers
		 * * `first_last_numbers` - 'First' and 'Last' buttons, plus page numbers
		 *  
		 * Further methods can be added using {@link DataTable.ext.oPagination}.
		 *  @type string
		 *  @default simple_numbers
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.pagingType
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "pagingType": "full_numbers"
		 *      } );
		 *    } )
		 */
		"sPaginationType": "simple_numbers",
	
	
		/**
		 * Enable horizontal scrolling. When a table is too wide to fit into a
		 * certain layout, or you have a large number of columns in the table, you
		 * can enable x-scrolling to show the table in a viewport, which can be
		 * scrolled. This property can be `true` which will allow the table to
		 * scroll horizontally when needed, or any CSS unit, or a number (in which
		 * case it will be treated as a pixel measurement). Setting as simply `true`
		 * is recommended.
		 *  @type boolean|string
		 *  @default <i>blank string - i.e. disabled</i>
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.scrollX
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "scrollX": true,
		 *        "scrollCollapse": true
		 *      } );
		 *    } );
		 */
		"sScrollX": "",
	
	
		/**
		 * This property can be used to force a DataTable to use more width than it
		 * might otherwise do when x-scrolling is enabled. For example if you have a
		 * table which requires to be well spaced, this parameter is useful for
		 * "over-sizing" the table, and thus forcing scrolling. This property can by
		 * any CSS unit, or a number (in which case it will be treated as a pixel
		 * measurement).
		 *  @type string
		 *  @default <i>blank string - i.e. disabled</i>
		 *
		 *  @dtopt Options
		 *  @name DataTable.defaults.scrollXInner
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "scrollX": "100%",
		 *        "scrollXInner": "110%"
		 *      } );
		 *    } );
		 */
		"sScrollXInner": "",
	
	
		/**
		 * Enable vertical scrolling. Vertical scrolling will constrain the DataTable
		 * to the given height, and enable scrolling for any data which overflows the
		 * current viewport. This can be used as an alternative to paging to display
		 * a lot of data in a small area (although paging and scrolling can both be
		 * enabled at the same time). This property can be any CSS unit, or a number
		 * (in which case it will be treated as a pixel measurement).
		 *  @type string
		 *  @default <i>blank string - i.e. disabled</i>
		 *
		 *  @dtopt Features
		 *  @name DataTable.defaults.scrollY
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "scrollY": "200px",
		 *        "paginate": false
		 *      } );
		 *    } );
		 */
		"sScrollY": "",
	
	
		/**
		 * __Deprecated__ The functionality provided by this parameter has now been
		 * superseded by that provided through `ajax`, which should be used instead.
		 *
		 * Set the HTTP method that is used to make the Ajax call for server-side
		 * processing or Ajax sourced data.
		 *  @type string
		 *  @default GET
		 *
		 *  @dtopt Options
		 *  @dtopt Server-side
		 *  @name DataTable.defaults.serverMethod
		 *
		 *  @deprecated 1.10. Please use `ajax` for this functionality now.
		 */
		"sServerMethod": "GET",
	
	
		/**
		 * DataTables makes use of renderers when displaying HTML elements for
		 * a table. These renderers can be added or modified by plug-ins to
		 * generate suitable mark-up for a site. For example the Bootstrap
		 * integration plug-in for DataTables uses a paging button renderer to
		 * display pagination buttons in the mark-up required by Bootstrap.
		 *
		 * For further information about the renderers available see
		 * DataTable.ext.renderer
		 *  @type string|object
		 *  @default null
		 *
		 *  @name DataTable.defaults.renderer
		 *
		 */
		"renderer": null,
	
	
		/**
		 * Set the data property name that DataTables should use to get a row's id
		 * to set as the `id` property in the node.
		 *  @type string
		 *  @default DT_RowId
		 *
		 *  @name DataTable.defaults.rowId
		 */
		"rowId": "DT_RowId"
	};
	
	_fnHungarianMap( DataTable.defaults );
	
	
	
	/*
	 * Developer note - See note in model.defaults.js about the use of Hungarian
	 * notation and camel case.
	 */
	
	/**
	 * Column options that can be given to DataTables at initialisation time.
	 *  @namespace
	 */
	DataTable.defaults.column = {
		/**
		 * Define which column(s) an order will occur on for this column. This
		 * allows a column's ordering to take multiple columns into account when
		 * doing a sort or use the data from a different column. For example first
		 * name / last name columns make sense to do a multi-column sort over the
		 * two columns.
		 *  @type array|int
		 *  @default null <i>Takes the value of the column index automatically</i>
		 *
		 *  @name DataTable.defaults.column.orderData
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "orderData": [ 0, 1 ], "targets": [ 0 ] },
		 *          { "orderData": [ 1, 0 ], "targets": [ 1 ] },
		 *          { "orderData": 2, "targets": [ 2 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "orderData": [ 0, 1 ] },
		 *          { "orderData": [ 1, 0 ] },
		 *          { "orderData": 2 },
		 *          null,
		 *          null
		 *        ]
		 *      } );
		 *    } );
		 */
		"aDataSort": null,
		"iDataSort": -1,
	
	
		/**
		 * You can control the default ordering direction, and even alter the
		 * behaviour of the sort handler (i.e. only allow ascending ordering etc)
		 * using this parameter.
		 *  @type array
		 *  @default [ 'asc', 'desc' ]
		 *
		 *  @name DataTable.defaults.column.orderSequence
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "orderSequence": [ "asc" ], "targets": [ 1 ] },
		 *          { "orderSequence": [ "desc", "asc", "asc" ], "targets": [ 2 ] },
		 *          { "orderSequence": [ "desc" ], "targets": [ 3 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          null,
		 *          { "orderSequence": [ "asc" ] },
		 *          { "orderSequence": [ "desc", "asc", "asc" ] },
		 *          { "orderSequence": [ "desc" ] },
		 *          null
		 *        ]
		 *      } );
		 *    } );
		 */
		"asSorting": [ 'asc', 'desc' ],
	
	
		/**
		 * Enable or disable filtering on the data in this column.
		 *  @type boolean
		 *  @default true
		 *
		 *  @name DataTable.defaults.column.searchable
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "searchable": false, "targets": [ 0 ] }
		 *        ] } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "searchable": false },
		 *          null,
		 *          null,
		 *          null,
		 *          null
		 *        ] } );
		 *    } );
		 */
		"bSearchable": true,
	
	
		/**
		 * Enable or disable ordering on this column.
		 *  @type boolean
		 *  @default true
		 *
		 *  @name DataTable.defaults.column.orderable
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "orderable": false, "targets": [ 0 ] }
		 *        ] } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "orderable": false },
		 *          null,
		 *          null,
		 *          null,
		 *          null
		 *        ] } );
		 *    } );
		 */
		"bSortable": true,
	
	
		/**
		 * Enable or disable the display of this column.
		 *  @type boolean
		 *  @default true
		 *
		 *  @name DataTable.defaults.column.visible
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "visible": false, "targets": [ 0 ] }
		 *        ] } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "visible": false },
		 *          null,
		 *          null,
		 *          null,
		 *          null
		 *        ] } );
		 *    } );
		 */
		"bVisible": true,
	
	
		/**
		 * Developer definable function that is called whenever a cell is created (Ajax source,
		 * etc) or processed for input (DOM source). This can be used as a compliment to mRender
		 * allowing you to modify the DOM element (add background colour for example) when the
		 * element is available.
		 *  @type function
		 *  @param {element} td The TD node that has been created
		 *  @param {*} cellData The Data for the cell
		 *  @param {array|object} rowData The data for the whole row
		 *  @param {int} row The row index for the aoData data store
		 *  @param {int} col The column index for aoColumns
		 *
		 *  @name DataTable.defaults.column.createdCell
		 *  @dtopt Columns
		 *
		 *  @example
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [3],
		 *          "createdCell": function (td, cellData, rowData, row, col) {
		 *            if ( cellData == "1.7" ) {
		 *              $(td).css('color', 'blue')
		 *            }
		 *          }
		 *        } ]
		 *      });
		 *    } );
		 */
		"fnCreatedCell": null,
	
	
		/**
		 * This parameter has been replaced by `data` in DataTables to ensure naming
		 * consistency. `dataProp` can still be used, as there is backwards
		 * compatibility in DataTables for this option, but it is strongly
		 * recommended that you use `data` in preference to `dataProp`.
		 *  @name DataTable.defaults.column.dataProp
		 */
	
	
		/**
		 * This property can be used to read data from any data source property,
		 * including deeply nested objects / properties. `data` can be given in a
		 * number of different ways which effect its behaviour:
		 *
		 * * `integer` - treated as an array index for the data source. This is the
		 *   default that DataTables uses (incrementally increased for each column).
		 * * `string` - read an object property from the data source. There are
		 *   three 'special' options that can be used in the string to alter how
		 *   DataTables reads the data from the source object:
		 *    * `.` - Dotted Javascript notation. Just as you use a `.` in
		 *      Javascript to read from nested objects, so to can the options
		 *      specified in `data`. For example: `browser.version` or
		 *      `browser.name`. If your object parameter name contains a period, use
		 *      `\\` to escape it - i.e. `first\\.name`.
		 *    * `[]` - Array notation. DataTables can automatically combine data
		 *      from and array source, joining the data with the characters provided
		 *      between the two brackets. For example: `name[, ]` would provide a
		 *      comma-space separated list from the source array. If no characters
		 *      are provided between the brackets, the original array source is
		 *      returned.
		 *    * `()` - Function notation. Adding `()` to the end of a parameter will
		 *      execute a function of the name given. For example: `browser()` for a
		 *      simple function on the data source, `browser.version()` for a
		 *      function in a nested property or even `browser().version` to get an
		 *      object property if the function called returns an object. Note that
		 *      function notation is recommended for use in `render` rather than
		 *      `data` as it is much simpler to use as a renderer.
		 * * `null` - use the original data source for the row rather than plucking
		 *   data directly from it. This action has effects on two other
		 *   initialisation options:
		 *    * `defaultContent` - When null is given as the `data` option and
		 *      `defaultContent` is specified for the column, the value defined by
		 *      `defaultContent` will be used for the cell.
		 *    * `render` - When null is used for the `data` option and the `render`
		 *      option is specified for the column, the whole data source for the
		 *      row is used for the renderer.
		 * * `function` - the function given will be executed whenever DataTables
		 *   needs to set or get the data for a cell in the column. The function
		 *   takes three parameters:
		 *    * Parameters:
		 *      * `{array|object}` The data source for the row
		 *      * `{string}` The type call data requested - this will be 'set' when
		 *        setting data or 'filter', 'display', 'type', 'sort' or undefined
		 *        when gathering data. Note that when `undefined` is given for the
		 *        type DataTables expects to get the raw data for the object back<
		 *      * `{*}` Data to set when the second parameter is 'set'.
		 *    * Return:
		 *      * The return value from the function is not required when 'set' is
		 *        the type of call, but otherwise the return is what will be used
		 *        for the data requested.
		 *
		 * Note that `data` is a getter and setter option. If you just require
		 * formatting of data for output, you will likely want to use `render` which
		 * is simply a getter and thus simpler to use.
		 *
		 * Note that prior to DataTables 1.9.2 `data` was called `mDataProp`. The
		 * name change reflects the flexibility of this property and is consistent
		 * with the naming of mRender. If 'mDataProp' is given, then it will still
		 * be used by DataTables, as it automatically maps the old name to the new
		 * if required.
		 *
		 *  @type string|int|function|null
		 *  @default null <i>Use automatically calculated column index</i>
		 *
		 *  @name DataTable.defaults.column.data
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Read table data from objects
		 *    // JSON structure for each row:
		 *    //   {
		 *    //      "engine": {value},
		 *    //      "browser": {value},
		 *    //      "platform": {value},
		 *    //      "version": {value},
		 *    //      "grade": {value}
		 *    //   }
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "ajaxSource": "sources/objects.txt",
		 *        "columns": [
		 *          { "data": "engine" },
		 *          { "data": "browser" },
		 *          { "data": "platform" },
		 *          { "data": "version" },
		 *          { "data": "grade" }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Read information from deeply nested objects
		 *    // JSON structure for each row:
		 *    //   {
		 *    //      "engine": {value},
		 *    //      "browser": {value},
		 *    //      "platform": {
		 *    //         "inner": {value}
		 *    //      },
		 *    //      "details": [
		 *    //         {value}, {value}
		 *    //      ]
		 *    //   }
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "ajaxSource": "sources/deep.txt",
		 *        "columns": [
		 *          { "data": "engine" },
		 *          { "data": "browser" },
		 *          { "data": "platform.inner" },
		 *          { "data": "details.0" },
		 *          { "data": "details.1" }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `data` as a function to provide different information for
		 *    // sorting, filtering and display. In this case, currency (price)
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [ 0 ],
		 *          "data": function ( source, type, val ) {
		 *            if (type === 'set') {
		 *              source.price = val;
		 *              // Store the computed display and filter values for efficiency
		 *              source.price_display = val=="" ? "" : "$"+numberFormat(val);
		 *              source.price_filter  = val=="" ? "" : "$"+numberFormat(val)+" "+val;
		 *              return;
		 *            }
		 *            else if (type === 'display') {
		 *              return source.price_display;
		 *            }
		 *            else if (type === 'filter') {
		 *              return source.price_filter;
		 *            }
		 *            // 'sort', 'type' and undefined all just use the integer
		 *            return source.price;
		 *          }
		 *        } ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using default content
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [ 0 ],
		 *          "data": null,
		 *          "defaultContent": "Click to edit"
		 *        } ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using array notation - outputting a list from an array
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [ 0 ],
		 *          "data": "name[, ]"
		 *        } ]
		 *      } );
		 *    } );
		 *
		 */
		"mData": null,
	
	
		/**
		 * This property is the rendering partner to `data` and it is suggested that
		 * when you want to manipulate data for display (including filtering,
		 * sorting etc) without altering the underlying data for the table, use this
		 * property. `render` can be considered to be the the read only companion to
		 * `data` which is read / write (then as such more complex). Like `data`
		 * this option can be given in a number of different ways to effect its
		 * behaviour:
		 *
		 * * `integer` - treated as an array index for the data source. This is the
		 *   default that DataTables uses (incrementally increased for each column).
		 * * `string` - read an object property from the data source. There are
		 *   three 'special' options that can be used in the string to alter how
		 *   DataTables reads the data from the source object:
		 *    * `.` - Dotted Javascript notation. Just as you use a `.` in
		 *      Javascript to read from nested objects, so to can the options
		 *      specified in `data`. For example: `browser.version` or
		 *      `browser.name`. If your object parameter name contains a period, use
		 *      `\\` to escape it - i.e. `first\\.name`.
		 *    * `[]` - Array notation. DataTables can automatically combine data
		 *      from and array source, joining the data with the characters provided
		 *      between the two brackets. For example: `name[, ]` would provide a
		 *      comma-space separated list from the source array. If no characters
		 *      are provided between the brackets, the original array source is
		 *      returned.
		 *    * `()` - Function notation. Adding `()` to the end of a parameter will
		 *      execute a function of the name given. For example: `browser()` for a
		 *      simple function on the data source, `browser.version()` for a
		 *      function in a nested property or even `browser().version` to get an
		 *      object property if the function called returns an object.
		 * * `object` - use different data for the different data types requested by
		 *   DataTables ('filter', 'display', 'type' or 'sort'). The property names
		 *   of the object is the data type the property refers to and the value can
		 *   defined using an integer, string or function using the same rules as
		 *   `render` normally does. Note that an `_` option _must_ be specified.
		 *   This is the default value to use if you haven't specified a value for
		 *   the data type requested by DataTables.
		 * * `function` - the function given will be executed whenever DataTables
		 *   needs to set or get the data for a cell in the column. The function
		 *   takes three parameters:
		 *    * Parameters:
		 *      * {array|object} The data source for the row (based on `data`)
		 *      * {string} The type call data requested - this will be 'filter',
		 *        'display', 'type' or 'sort'.
		 *      * {array|object} The full data source for the row (not based on
		 *        `data`)
		 *    * Return:
		 *      * The return value from the function is what will be used for the
		 *        data requested.
		 *
		 *  @type string|int|function|object|null
		 *  @default null Use the data source value.
		 *
		 *  @name DataTable.defaults.column.render
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Create a comma separated list from an array of objects
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "ajaxSource": "sources/deep.txt",
		 *        "columns": [
		 *          { "data": "engine" },
		 *          { "data": "browser" },
		 *          {
		 *            "data": "platform",
		 *            "render": "[, ].name"
		 *          }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Execute a function to obtain data
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [ 0 ],
		 *          "data": null, // Use the full data source object for the renderer's source
		 *          "render": "browserName()"
		 *        } ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // As an object, extracting different data for the different types
		 *    // This would be used with a data source such as:
		 *    //   { "phone": 5552368, "phone_filter": "5552368 555-2368", "phone_display": "555-2368" }
		 *    // Here the `phone` integer is used for sorting and type detection, while `phone_filter`
		 *    // (which has both forms) is used for filtering for if a user inputs either format, while
		 *    // the formatted phone number is the one that is shown in the table.
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [ 0 ],
		 *          "data": null, // Use the full data source object for the renderer's source
		 *          "render": {
		 *            "_": "phone",
		 *            "filter": "phone_filter",
		 *            "display": "phone_display"
		 *          }
		 *        } ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Use as a function to create a link from the data source
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [ 0 ],
		 *          "data": "download_link",
		 *          "render": function ( data, type, full ) {
		 *            return '<a href="'+data+'">Download</a>';
		 *          }
		 *        } ]
		 *      } );
		 *    } );
		 */
		"mRender": null,
	
	
		/**
		 * Change the cell type created for the column - either TD cells or TH cells. This
		 * can be useful as TH cells have semantic meaning in the table body, allowing them
		 * to act as a header for a row (you may wish to add scope='row' to the TH elements).
		 *  @type string
		 *  @default td
		 *
		 *  @name DataTable.defaults.column.cellType
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Make the first column use TH cells
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [ {
		 *          "targets": [ 0 ],
		 *          "cellType": "th"
		 *        } ]
		 *      } );
		 *    } );
		 */
		"sCellType": "td",
	
	
		/**
		 * Class to give to each cell in this column.
		 *  @type string
		 *  @default <i>Empty string</i>
		 *
		 *  @name DataTable.defaults.column.class
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "class": "my_class", "targets": [ 0 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "class": "my_class" },
		 *          null,
		 *          null,
		 *          null,
		 *          null
		 *        ]
		 *      } );
		 *    } );
		 */
		"sClass": "",
	
		/**
		 * When DataTables calculates the column widths to assign to each column,
		 * it finds the longest string in each column and then constructs a
		 * temporary table and reads the widths from that. The problem with this
		 * is that "mmm" is much wider then "iiii", but the latter is a longer
		 * string - thus the calculation can go wrong (doing it properly and putting
		 * it into an DOM object and measuring that is horribly(!) slow). Thus as
		 * a "work around" we provide this option. It will append its value to the
		 * text that is found to be the longest string for the column - i.e. padding.
		 * Generally you shouldn't need this!
		 *  @type string
		 *  @default <i>Empty string<i>
		 *
		 *  @name DataTable.defaults.column.contentPadding
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          null,
		 *          null,
		 *          null,
		 *          {
		 *            "contentPadding": "mmm"
		 *          }
		 *        ]
		 *      } );
		 *    } );
		 */
		"sContentPadding": "",
	
	
		/**
		 * Allows a default value to be given for a column's data, and will be used
		 * whenever a null data source is encountered (this can be because `data`
		 * is set to null, or because the data source itself is null).
		 *  @type string
		 *  @default null
		 *
		 *  @name DataTable.defaults.column.defaultContent
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          {
		 *            "data": null,
		 *            "defaultContent": "Edit",
		 *            "targets": [ -1 ]
		 *          }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          null,
		 *          null,
		 *          null,
		 *          {
		 *            "data": null,
		 *            "defaultContent": "Edit"
		 *          }
		 *        ]
		 *      } );
		 *    } );
		 */
		"sDefaultContent": null,
	
	
		/**
		 * This parameter is only used in DataTables' server-side processing. It can
		 * be exceptionally useful to know what columns are being displayed on the
		 * client side, and to map these to database fields. When defined, the names
		 * also allow DataTables to reorder information from the server if it comes
		 * back in an unexpected order (i.e. if you switch your columns around on the
		 * client-side, your server-side code does not also need updating).
		 *  @type string
		 *  @default <i>Empty string</i>
		 *
		 *  @name DataTable.defaults.column.name
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "name": "engine", "targets": [ 0 ] },
		 *          { "name": "browser", "targets": [ 1 ] },
		 *          { "name": "platform", "targets": [ 2 ] },
		 *          { "name": "version", "targets": [ 3 ] },
		 *          { "name": "grade", "targets": [ 4 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "name": "engine" },
		 *          { "name": "browser" },
		 *          { "name": "platform" },
		 *          { "name": "version" },
		 *          { "name": "grade" }
		 *        ]
		 *      } );
		 *    } );
		 */
		"sName": "",
	
	
		/**
		 * Defines a data source type for the ordering which can be used to read
		 * real-time information from the table (updating the internally cached
		 * version) prior to ordering. This allows ordering to occur on user
		 * editable elements such as form inputs.
		 *  @type string
		 *  @default std
		 *
		 *  @name DataTable.defaults.column.orderDataType
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "orderDataType": "dom-text", "targets": [ 2, 3 ] },
		 *          { "type": "numeric", "targets": [ 3 ] },
		 *          { "orderDataType": "dom-select", "targets": [ 4 ] },
		 *          { "orderDataType": "dom-checkbox", "targets": [ 5 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          null,
		 *          null,
		 *          { "orderDataType": "dom-text" },
		 *          { "orderDataType": "dom-text", "type": "numeric" },
		 *          { "orderDataType": "dom-select" },
		 *          { "orderDataType": "dom-checkbox" }
		 *        ]
		 *      } );
		 *    } );
		 */
		"sSortDataType": "std",
	
	
		/**
		 * The title of this column.
		 *  @type string
		 *  @default null <i>Derived from the 'TH' value for this column in the
		 *    original HTML table.</i>
		 *
		 *  @name DataTable.defaults.column.title
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "title": "My column title", "targets": [ 0 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "title": "My column title" },
		 *          null,
		 *          null,
		 *          null,
		 *          null
		 *        ]
		 *      } );
		 *    } );
		 */
		"sTitle": null,
	
	
		/**
		 * The type allows you to specify how the data for this column will be
		 * ordered. Four types (string, numeric, date and html (which will strip
		 * HTML tags before ordering)) are currently available. Note that only date
		 * formats understood by Javascript's Date() object will be accepted as type
		 * date. For example: "Mar 26, 2008 5:03 PM". May take the values: 'string',
		 * 'numeric', 'date' or 'html' (by default). Further types can be adding
		 * through plug-ins.
		 *  @type string
		 *  @default null <i>Auto-detected from raw data</i>
		 *
		 *  @name DataTable.defaults.column.type
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "type": "html", "targets": [ 0 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "type": "html" },
		 *          null,
		 *          null,
		 *          null,
		 *          null
		 *        ]
		 *      } );
		 *    } );
		 */
		"sType": null,
	
	
		/**
		 * Defining the width of the column, this parameter may take any CSS value
		 * (3em, 20px etc). DataTables applies 'smart' widths to columns which have not
		 * been given a specific width through this interface ensuring that the table
		 * remains readable.
		 *  @type string
		 *  @default null <i>Automatic</i>
		 *
		 *  @name DataTable.defaults.column.width
		 *  @dtopt Columns
		 *
		 *  @example
		 *    // Using `columnDefs`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columnDefs": [
		 *          { "width": "20%", "targets": [ 0 ] }
		 *        ]
		 *      } );
		 *    } );
		 *
		 *  @example
		 *    // Using `columns`
		 *    $(document).ready( function() {
		 *      $('#example').dataTable( {
		 *        "columns": [
		 *          { "width": "20%" },
		 *          null,
		 *          null,
		 *          null,
		 *          null
		 *        ]
		 *      } );
		 *    } );
		 */
		"sWidth": null
	};
	
	_fnHungarianMap( DataTable.defaults.column );
	
	
	
	/**
	 * DataTables settings object - this holds all the information needed for a
	 * given table, including configuration, data and current application of the
	 * table options. DataTables does not have a single instance for each DataTable
	 * with the settings attached to that instance, but rather instances of the
	 * DataTable "class" are created on-the-fly as needed (typically by a
	 * $().dataTable() call) and the settings object is then applied to that
	 * instance.
	 *
	 * Note that this object is related to {@link DataTable.defaults} but this
	 * one is the internal data store for DataTables's cache of columns. It should
	 * NOT be manipulated outside of DataTables. Any configuration should be done
	 * through the initialisation options.
	 *  @namespace
	 *  @todo Really should attach the settings object to individual instances so we
	 *    don't need to create new instances on each $().dataTable() call (if the
	 *    table already exists). It would also save passing oSettings around and
	 *    into every single function. However, this is a very significant
	 *    architecture change for DataTables and will almost certainly break
	 *    backwards compatibility with older installations. This is something that
	 *    will be done in 2.0.
	 */
	DataTable.models.oSettings = {
		/**
		 * Primary features of DataTables and their enablement state.
		 *  @namespace
		 */
		"oFeatures": {
	
			/**
			 * Flag to say if DataTables should automatically try to calculate the
			 * optimum table and columns widths (true) or not (false).
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bAutoWidth": null,
	
			/**
			 * Delay the creation of TR and TD elements until they are actually
			 * needed by a driven page draw. This can give a significant speed
			 * increase for Ajax source and Javascript source data, but makes no
			 * difference at all for DOM and server-side processing tables.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bDeferRender": null,
	
			/**
			 * Enable filtering on the table or not. Note that if this is disabled
			 * then there is no filtering at all on the table, including fnFilter.
			 * To just remove the filtering input use sDom and remove the 'f' option.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bFilter": null,
	
			/**
			 * Table information element (the 'Showing x of y records' div) enable
			 * flag.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bInfo": null,
	
			/**
			 * Present a user control allowing the end user to change the page size
			 * when pagination is enabled.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bLengthChange": null,
	
			/**
			 * Pagination enabled or not. Note that if this is disabled then length
			 * changing must also be disabled.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bPaginate": null,
	
			/**
			 * Processing indicator enable flag whenever DataTables is enacting a
			 * user request - typically an Ajax request for server-side processing.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bProcessing": null,
	
			/**
			 * Server-side processing enabled flag - when enabled DataTables will
			 * get all data from the server for every draw - there is no filtering,
			 * sorting or paging done on the client-side.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bServerSide": null,
	
			/**
			 * Sorting enablement flag.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bSort": null,
	
			/**
			 * Multi-column sorting
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bSortMulti": null,
	
			/**
			 * Apply a class to the columns which are being sorted to provide a
			 * visual highlight or not. This can slow things down when enabled since
			 * there is a lot of DOM interaction.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bSortClasses": null,
	
			/**
			 * State saving enablement flag.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bStateSave": null
		},
	
	
		/**
		 * Scrolling settings for a table.
		 *  @namespace
		 */
		"oScroll": {
			/**
			 * When the table is shorter in height than sScrollY, collapse the
			 * table container down to the height of the table (when true).
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type boolean
			 */
			"bCollapse": null,
	
			/**
			 * Width of the scrollbar for the web-browser's platform. Calculated
			 * during table initialisation.
			 *  @type int
			 *  @default 0
			 */
			"iBarWidth": 0,
	
			/**
			 * Viewport width for horizontal scrolling. Horizontal scrolling is
			 * disabled if an empty string.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type string
			 */
			"sX": null,
	
			/**
			 * Width to expand the table to when using x-scrolling. Typically you
			 * should not need to use this.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type string
			 *  @deprecated
			 */
			"sXInner": null,
	
			/**
			 * Viewport height for vertical scrolling. Vertical scrolling is disabled
			 * if an empty string.
			 * Note that this parameter will be set by the initialisation routine. To
			 * set a default use {@link DataTable.defaults}.
			 *  @type string
			 */
			"sY": null
		},
	
		/**
		 * Language information for the table.
		 *  @namespace
		 *  @extends DataTable.defaults.oLanguage
		 */
		"oLanguage": {
			/**
			 * Information callback function. See
			 * {@link DataTable.defaults.fnInfoCallback}
			 *  @type function
			 *  @default null
			 */
			"fnInfoCallback": null
		},
	
		/**
		 * Browser support parameters
		 *  @namespace
		 */
		"oBrowser": {
			/**
			 * Indicate if the browser incorrectly calculates width:100% inside a
			 * scrolling element (IE6/7)
			 *  @type boolean
			 *  @default false
			 */
			"bScrollOversize": false,
	
			/**
			 * Determine if the vertical scrollbar is on the right or left of the
			 * scrolling container - needed for rtl language layout, although not
			 * all browsers move the scrollbar (Safari).
			 *  @type boolean
			 *  @default false
			 */
			"bScrollbarLeft": false,
	
			/**
			 * Flag for if `getBoundingClientRect` is fully supported or not
			 *  @type boolean
			 *  @default false
			 */
			"bBounding": false,
	
			/**
			 * Browser scrollbar width
			 *  @type integer
			 *  @default 0
			 */
			"barWidth": 0
		},
	
	
		"ajax": null,
	
	
		/**
		 * Array referencing the nodes which are used for the features. The
		 * parameters of this object match what is allowed by sDom - i.e.
		 *   <ul>
		 *     <li>'l' - Length changing</li>
		 *     <li>'f' - Filtering input</li>
		 *     <li>'t' - The table!</li>
		 *     <li>'i' - Information</li>
		 *     <li>'p' - Pagination</li>
		 *     <li>'r' - pRocessing</li>
		 *   </ul>
		 *  @type array
		 *  @default []
		 */
		"aanFeatures": [],
	
		/**
		 * Store data information - see {@link DataTable.models.oRow} for detailed
		 * information.
		 *  @type array
		 *  @default []
		 */
		"aoData": [],
	
		/**
		 * Array of indexes which are in the current display (after filtering etc)
		 *  @type array
		 *  @default []
		 */
		"aiDisplay": [],
	
		/**
		 * Array of indexes for display - no filtering
		 *  @type array
		 *  @default []
		 */
		"aiDisplayMaster": [],
	
		/**
		 * Map of row ids to data indexes
		 *  @type object
		 *  @default {}
		 */
		"aIds": {},
	
		/**
		 * Store information about each column that is in use
		 *  @type array
		 *  @default []
		 */
		"aoColumns": [],
	
		/**
		 * Store information about the table's header
		 *  @type array
		 *  @default []
		 */
		"aoHeader": [],
	
		/**
		 * Store information about the table's footer
		 *  @type array
		 *  @default []
		 */
		"aoFooter": [],
	
		/**
		 * Store the applied global search information in case we want to force a
		 * research or compare the old search to a new one.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @namespace
		 *  @extends DataTable.models.oSearch
		 */
		"oPreviousSearch": {},
	
		/**
		 * Store the applied search for each column - see
		 * {@link DataTable.models.oSearch} for the format that is used for the
		 * filtering information for each column.
		 *  @type array
		 *  @default []
		 */
		"aoPreSearchCols": [],
	
		/**
		 * Sorting that is applied to the table. Note that the inner arrays are
		 * used in the following manner:
		 * <ul>
		 *   <li>Index 0 - column number</li>
		 *   <li>Index 1 - current sorting direction</li>
		 * </ul>
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type array
		 *  @todo These inner arrays should really be objects
		 */
		"aaSorting": null,
	
		/**
		 * Sorting that is always applied to the table (i.e. prefixed in front of
		 * aaSorting).
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type array
		 *  @default []
		 */
		"aaSortingFixed": [],
	
		/**
		 * Classes to use for the striping of a table.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type array
		 *  @default []
		 */
		"asStripeClasses": null,
	
		/**
		 * If restoring a table - we should restore its striping classes as well
		 *  @type array
		 *  @default []
		 */
		"asDestroyStripes": [],
	
		/**
		 * If restoring a table - we should restore its width
		 *  @type int
		 *  @default 0
		 */
		"sDestroyWidth": 0,
	
		/**
		 * Callback functions array for every time a row is inserted (i.e. on a draw).
		 *  @type array
		 *  @default []
		 */
		"aoRowCallback": [],
	
		/**
		 * Callback functions for the header on each draw.
		 *  @type array
		 *  @default []
		 */
		"aoHeaderCallback": [],
	
		/**
		 * Callback function for the footer on each draw.
		 *  @type array
		 *  @default []
		 */
		"aoFooterCallback": [],
	
		/**
		 * Array of callback functions for draw callback functions
		 *  @type array
		 *  @default []
		 */
		"aoDrawCallback": [],
	
		/**
		 * Array of callback functions for row created function
		 *  @type array
		 *  @default []
		 */
		"aoRowCreatedCallback": [],
	
		/**
		 * Callback functions for just before the table is redrawn. A return of
		 * false will be used to cancel the draw.
		 *  @type array
		 *  @default []
		 */
		"aoPreDrawCallback": [],
	
		/**
		 * Callback functions for when the table has been initialised.
		 *  @type array
		 *  @default []
		 */
		"aoInitComplete": [],
	
	
		/**
		 * Callbacks for modifying the settings to be stored for state saving, prior to
		 * saving state.
		 *  @type array
		 *  @default []
		 */
		"aoStateSaveParams": [],
	
		/**
		 * Callbacks for modifying the settings that have been stored for state saving
		 * prior to using the stored values to restore the state.
		 *  @type array
		 *  @default []
		 */
		"aoStateLoadParams": [],
	
		/**
		 * Callbacks for operating on the settings object once the saved state has been
		 * loaded
		 *  @type array
		 *  @default []
		 */
		"aoStateLoaded": [],
	
		/**
		 * Cache the table ID for quick access
		 *  @type string
		 *  @default <i>Empty string</i>
		 */
		"sTableId": "",
	
		/**
		 * The TABLE node for the main table
		 *  @type node
		 *  @default null
		 */
		"nTable": null,
	
		/**
		 * Permanent ref to the thead element
		 *  @type node
		 *  @default null
		 */
		"nTHead": null,
	
		/**
		 * Permanent ref to the tfoot element - if it exists
		 *  @type node
		 *  @default null
		 */
		"nTFoot": null,
	
		/**
		 * Permanent ref to the tbody element
		 *  @type node
		 *  @default null
		 */
		"nTBody": null,
	
		/**
		 * Cache the wrapper node (contains all DataTables controlled elements)
		 *  @type node
		 *  @default null
		 */
		"nTableWrapper": null,
	
		/**
		 * Indicate if when using server-side processing the loading of data
		 * should be deferred until the second draw.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type boolean
		 *  @default false
		 */
		"bDeferLoading": false,
	
		/**
		 * Indicate if all required information has been read in
		 *  @type boolean
		 *  @default false
		 */
		"bInitialised": false,
	
		/**
		 * Information about open rows. Each object in the array has the parameters
		 * 'nTr' and 'nParent'
		 *  @type array
		 *  @default []
		 */
		"aoOpenRows": [],
	
		/**
		 * Dictate the positioning of DataTables' control elements - see
		 * {@link DataTable.model.oInit.sDom}.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type string
		 *  @default null
		 */
		"sDom": null,
	
		/**
		 * Search delay (in mS)
		 *  @type integer
		 *  @default null
		 */
		"searchDelay": null,
	
		/**
		 * Which type of pagination should be used.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type string
		 *  @default two_button
		 */
		"sPaginationType": "two_button",
	
		/**
		 * The state duration (for `stateSave`) in seconds.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type int
		 *  @default 0
		 */
		"iStateDuration": 0,
	
		/**
		 * Array of callback functions for state saving. Each array element is an
		 * object with the following parameters:
		 *   <ul>
		 *     <li>function:fn - function to call. Takes two parameters, oSettings
		 *       and the JSON string to save that has been thus far created. Returns
		 *       a JSON string to be inserted into a json object
		 *       (i.e. '"param": [ 0, 1, 2]')</li>
		 *     <li>string:sName - name of callback</li>
		 *   </ul>
		 *  @type array
		 *  @default []
		 */
		"aoStateSave": [],
	
		/**
		 * Array of callback functions for state loading. Each array element is an
		 * object with the following parameters:
		 *   <ul>
		 *     <li>function:fn - function to call. Takes two parameters, oSettings
		 *       and the object stored. May return false to cancel state loading</li>
		 *     <li>string:sName - name of callback</li>
		 *   </ul>
		 *  @type array
		 *  @default []
		 */
		"aoStateLoad": [],
	
		/**
		 * State that was saved. Useful for back reference
		 *  @type object
		 *  @default null
		 */
		"oSavedState": null,
	
		/**
		 * State that was loaded. Useful for back reference
		 *  @type object
		 *  @default null
		 */
		"oLoadedState": null,
	
		/**
		 * Source url for AJAX data for the table.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type string
		 *  @default null
		 */
		"sAjaxSource": null,
	
		/**
		 * Property from a given object from which to read the table data from. This
		 * can be an empty string (when not server-side processing), in which case
		 * it is  assumed an an array is given directly.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type string
		 */
		"sAjaxDataProp": null,
	
		/**
		 * The last jQuery XHR object that was used for server-side data gathering.
		 * This can be used for working with the XHR information in one of the
		 * callbacks
		 *  @type object
		 *  @default null
		 */
		"jqXHR": null,
	
		/**
		 * JSON returned from the server in the last Ajax request
		 *  @type object
		 *  @default undefined
		 */
		"json": undefined,
	
		/**
		 * Data submitted as part of the last Ajax request
		 *  @type object
		 *  @default undefined
		 */
		"oAjaxData": undefined,
	
		/**
		 * Function to get the server-side data.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type function
		 */
		"fnServerData": null,
	
		/**
		 * Functions which are called prior to sending an Ajax request so extra
		 * parameters can easily be sent to the server
		 *  @type array
		 *  @default []
		 */
		"aoServerParams": [],
	
		/**
		 * Send the XHR HTTP method - GET or POST (could be PUT or DELETE if
		 * required).
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type string
		 */
		"sServerMethod": null,
	
		/**
		 * Format numbers for display.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type function
		 */
		"fnFormatNumber": null,
	
		/**
		 * List of options that can be used for the user selectable length menu.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type array
		 *  @default []
		 */
		"aLengthMenu": null,
	
		/**
		 * Counter for the draws that the table does. Also used as a tracker for
		 * server-side processing
		 *  @type int
		 *  @default 0
		 */
		"iDraw": 0,
	
		/**
		 * Indicate if a redraw is being done - useful for Ajax
		 *  @type boolean
		 *  @default false
		 */
		"bDrawing": false,
	
		/**
		 * Draw index (iDraw) of the last error when parsing the returned data
		 *  @type int
		 *  @default -1
		 */
		"iDrawError": -1,
	
		/**
		 * Paging display length
		 *  @type int
		 *  @default 10
		 */
		"_iDisplayLength": 10,
	
		/**
		 * Paging start point - aiDisplay index
		 *  @type int
		 *  @default 0
		 */
		"_iDisplayStart": 0,
	
		/**
		 * Server-side processing - number of records in the result set
		 * (i.e. before filtering), Use fnRecordsTotal rather than
		 * this property to get the value of the number of records, regardless of
		 * the server-side processing setting.
		 *  @type int
		 *  @default 0
		 *  @private
		 */
		"_iRecordsTotal": 0,
	
		/**
		 * Server-side processing - number of records in the current display set
		 * (i.e. after filtering). Use fnRecordsDisplay rather than
		 * this property to get the value of the number of records, regardless of
		 * the server-side processing setting.
		 *  @type boolean
		 *  @default 0
		 *  @private
		 */
		"_iRecordsDisplay": 0,
	
		/**
		 * The classes to use for the table
		 *  @type object
		 *  @default {}
		 */
		"oClasses": {},
	
		/**
		 * Flag attached to the settings object so you can check in the draw
		 * callback if filtering has been done in the draw. Deprecated in favour of
		 * events.
		 *  @type boolean
		 *  @default false
		 *  @deprecated
		 */
		"bFiltered": false,
	
		/**
		 * Flag attached to the settings object so you can check in the draw
		 * callback if sorting has been done in the draw. Deprecated in favour of
		 * events.
		 *  @type boolean
		 *  @default false
		 *  @deprecated
		 */
		"bSorted": false,
	
		/**
		 * Indicate that if multiple rows are in the header and there is more than
		 * one unique cell per column, if the top one (true) or bottom one (false)
		 * should be used for sorting / title by DataTables.
		 * Note that this parameter will be set by the initialisation routine. To
		 * set a default use {@link DataTable.defaults}.
		 *  @type boolean
		 */
		"bSortCellsTop": null,
	
		/**
		 * Initialisation object that is used for the table
		 *  @type object
		 *  @default null
		 */
		"oInit": null,
	
		/**
		 * Destroy callback functions - for plug-ins to attach themselves to the
		 * destroy so they can clean up markup and events.
		 *  @type array
		 *  @default []
		 */
		"aoDestroyCallback": [],
	
	
		/**
		 * Get the number of records in the current record set, before filtering
		 *  @type function
		 */
		"fnRecordsTotal": function ()
		{
			return _fnDataSource( this ) == 'ssp' ?
				this._iRecordsTotal * 1 :
				this.aiDisplayMaster.length;
		},
	
		/**
		 * Get the number of records in the current record set, after filtering
		 *  @type function
		 */
		"fnRecordsDisplay": function ()
		{
			return _fnDataSource( this ) == 'ssp' ?
				this._iRecordsDisplay * 1 :
				this.aiDisplay.length;
		},
	
		/**
		 * Get the display end point - aiDisplay index
		 *  @type function
		 */
		"fnDisplayEnd": function ()
		{
			var
				len      = this._iDisplayLength,
				start    = this._iDisplayStart,
				calc     = start + len,
				records  = this.aiDisplay.length,
				features = this.oFeatures,
				paginate = features.bPaginate;
	
			if ( features.bServerSide ) {
				return paginate === false || len === -1 ?
					start + records :
					Math.min( start+len, this._iRecordsDisplay );
			}
			else {
				return ! paginate || calc>records || len===-1 ?
					records :
					calc;
			}
		},
	
		/**
		 * The DataTables object for this table
		 *  @type object
		 *  @default null
		 */
		"oInstance": null,
	
		/**
		 * Unique identifier for each instance of the DataTables object. If there
		 * is an ID on the table node, then it takes that value, otherwise an
		 * incrementing internal counter is used.
		 *  @type string
		 *  @default null
		 */
		"sInstance": null,
	
		/**
		 * tabindex attribute value that is added to DataTables control elements, allowing
		 * keyboard navigation of the table and its controls.
		 */
		"iTabIndex": 0,
	
		/**
		 * DIV container for the footer scrolling table if scrolling
		 */
		"nScrollHead": null,
	
		/**
		 * DIV container for the footer scrolling table if scrolling
		 */
		"nScrollFoot": null,
	
		/**
		 * Last applied sort
		 *  @type array
		 *  @default []
		 */
		"aLastSort": [],
	
		/**
		 * Stored plug-in instances
		 *  @type object
		 *  @default {}
		 */
		"oPlugins": {},
	
		/**
		 * Function used to get a row's id from the row's data
		 *  @type function
		 *  @default null
		 */
		"rowIdFn": null,
	
		/**
		 * Data location where to store a row's id
		 *  @type string
		 *  @default null
		 */
		"rowId": null
	};
	
	/**
	 * Extension object for DataTables that is used to provide all extension
	 * options.
	 *
	 * Note that the `DataTable.ext` object is available through
	 * `jQuery.fn.dataTable.ext` where it may be accessed and manipulated. It is
	 * also aliased to `jQuery.fn.dataTableExt` for historic reasons.
	 *  @namespace
	 *  @extends DataTable.models.ext
	 */
	
	
	/**
	 * DataTables extensions
	 * 
	 * This namespace acts as a collection area for plug-ins that can be used to
	 * extend DataTables capabilities. Indeed many of the build in methods
	 * use this method to provide their own capabilities (sorting methods for
	 * example).
	 *
	 * Note that this namespace is aliased to `jQuery.fn.dataTableExt` for legacy
	 * reasons
	 *
	 *  @namespace
	 */
	DataTable.ext = _ext = {
		/**
		 * Buttons. For use with the Buttons extension for DataTables. This is
		 * defined here so other extensions can define buttons regardless of load
		 * order. It is _not_ used by DataTables core.
		 *
		 *  @type object
		 *  @default {}
		 */
		buttons: {},
	
	
		/**
		 * Element class names
		 *
		 *  @type object
		 *  @default {}
		 */
		classes: {},
	
	
		/**
		 * DataTables build type (expanded by the download builder)
		 *
		 *  @type string
		 */
		build:"ju/dt-1.13.4/e-2.1.3",
	
	
		/**
		 * Error reporting.
		 * 
		 * How should DataTables report an error. Can take the value 'alert',
		 * 'throw', 'none' or a function.
		 *
		 *  @type string|function
		 *  @default alert
		 */
		errMode: "alert",
	
	
		/**
		 * Feature plug-ins.
		 * 
		 * This is an array of objects which describe the feature plug-ins that are
		 * available to DataTables. These feature plug-ins are then available for
		 * use through the `dom` initialisation option.
		 * 
		 * Each feature plug-in is described by an object which must have the
		 * following properties:
		 * 
		 * * `fnInit` - function that is used to initialise the plug-in,
		 * * `cFeature` - a character so the feature can be enabled by the `dom`
		 *   instillation option. This is case sensitive.
		 *
		 * The `fnInit` function has the following input parameters:
		 *
		 * 1. `{object}` DataTables settings object: see
		 *    {@link DataTable.models.oSettings}
		 *
		 * And the following return is expected:
		 * 
		 * * {node|null} The element which contains your feature. Note that the
		 *   return may also be void if your plug-in does not require to inject any
		 *   DOM elements into DataTables control (`dom`) - for example this might
		 *   be useful when developing a plug-in which allows table control via
		 *   keyboard entry
		 *
		 *  @type array
		 *
		 *  @example
		 *    $.fn.dataTable.ext.features.push( {
		 *      "fnInit": function( oSettings ) {
		 *        return new TableTools( { "oDTSettings": oSettings } );
		 *      },
		 *      "cFeature": "T"
		 *    } );
		 */
		feature: [],
	
	
		/**
		 * Row searching.
		 * 
		 * This method of searching is complimentary to the default type based
		 * searching, and a lot more comprehensive as it allows you complete control
		 * over the searching logic. Each element in this array is a function
		 * (parameters described below) that is called for every row in the table,
		 * and your logic decides if it should be included in the searching data set
		 * or not.
		 *
		 * Searching functions have the following input parameters:
		 *
		 * 1. `{object}` DataTables settings object: see
		 *    {@link DataTable.models.oSettings}
		 * 2. `{array|object}` Data for the row to be processed (same as the
		 *    original format that was passed in as the data source, or an array
		 *    from a DOM data source
		 * 3. `{int}` Row index ({@link DataTable.models.oSettings.aoData}), which
		 *    can be useful to retrieve the `TR` element if you need DOM interaction.
		 *
		 * And the following return is expected:
		 *
		 * * {boolean} Include the row in the searched result set (true) or not
		 *   (false)
		 *
		 * Note that as with the main search ability in DataTables, technically this
		 * is "filtering", since it is subtractive. However, for consistency in
		 * naming we call it searching here.
		 *
		 *  @type array
		 *  @default []
		 *
		 *  @example
		 *    // The following example shows custom search being applied to the
		 *    // fourth column (i.e. the data[3] index) based on two input values
		 *    // from the end-user, matching the data in a certain range.
		 *    $.fn.dataTable.ext.search.push(
		 *      function( settings, data, dataIndex ) {
		 *        var min = document.getElementById('min').value * 1;
		 *        var max = document.getElementById('max').value * 1;
		 *        var version = data[3] == "-" ? 0 : data[3]*1;
		 *
		 *        if ( min == "" && max == "" ) {
		 *          return true;
		 *        }
		 *        else if ( min == "" && version < max ) {
		 *          return true;
		 *        }
		 *        else if ( min < version && "" == max ) {
		 *          return true;
		 *        }
		 *        else if ( min < version && version < max ) {
		 *          return true;
		 *        }
		 *        return false;
		 *      }
		 *    );
		 */
		search: [],
	
	
		/**
		 * Selector extensions
		 *
		 * The `selector` option can be used to extend the options available for the
		 * selector modifier options (`selector-modifier` object data type) that
		 * each of the three built in selector types offer (row, column and cell +
		 * their plural counterparts). For example the Select extension uses this
		 * mechanism to provide an option to select only rows, columns and cells
		 * that have been marked as selected by the end user (`{selected: true}`),
		 * which can be used in conjunction with the existing built in selector
		 * options.
		 *
		 * Each property is an array to which functions can be pushed. The functions
		 * take three attributes:
		 *
		 * * Settings object for the host table
		 * * Options object (`selector-modifier` object type)
		 * * Array of selected item indexes
		 *
		 * The return is an array of the resulting item indexes after the custom
		 * selector has been applied.
		 *
		 *  @type object
		 */
		selector: {
			cell: [],
			column: [],
			row: []
		},
	
	
		/**
		 * Internal functions, exposed for used in plug-ins.
		 * 
		 * Please note that you should not need to use the internal methods for
		 * anything other than a plug-in (and even then, try to avoid if possible).
		 * The internal function may change between releases.
		 *
		 *  @type object
		 *  @default {}
		 */
		internal: {},
	
	
		/**
		 * Legacy configuration options. Enable and disable legacy options that
		 * are available in DataTables.
		 *
		 *  @type object
		 */
		legacy: {
			/**
			 * Enable / disable DataTables 1.9 compatible server-side processing
			 * requests
			 *
			 *  @type boolean
			 *  @default null
			 */
			ajax: null
		},
	
	
		/**
		 * Pagination plug-in methods.
		 * 
		 * Each entry in this object is a function and defines which buttons should
		 * be shown by the pagination rendering method that is used for the table:
		 * {@link DataTable.ext.renderer.pageButton}. The renderer addresses how the
		 * buttons are displayed in the document, while the functions here tell it
		 * what buttons to display. This is done by returning an array of button
		 * descriptions (what each button will do).
		 *
		 * Pagination types (the four built in options and any additional plug-in
		 * options defined here) can be used through the `paginationType`
		 * initialisation parameter.
		 *
		 * The functions defined take two parameters:
		 *
		 * 1. `{int} page` The current page index
		 * 2. `{int} pages` The number of pages in the table
		 *
		 * Each function is expected to return an array where each element of the
		 * array can be one of:
		 *
		 * * `first` - Jump to first page when activated
		 * * `last` - Jump to last page when activated
		 * * `previous` - Show previous page when activated
		 * * `next` - Show next page when activated
		 * * `{int}` - Show page of the index given
		 * * `{array}` - A nested array containing the above elements to add a
		 *   containing 'DIV' element (might be useful for styling).
		 *
		 * Note that DataTables v1.9- used this object slightly differently whereby
		 * an object with two functions would be defined for each plug-in. That
		 * ability is still supported by DataTables 1.10+ to provide backwards
		 * compatibility, but this option of use is now decremented and no longer
		 * documented in DataTables 1.10+.
		 *
		 *  @type object
		 *  @default {}
		 *
		 *  @example
		 *    // Show previous, next and current page buttons only
		 *    $.fn.dataTableExt.oPagination.current = function ( page, pages ) {
		 *      return [ 'previous', page, 'next' ];
		 *    };
		 */
		pager: {},
	
	
		renderer: {
			pageButton: {},
			header: {}
		},
	
	
		/**
		 * Ordering plug-ins - custom data source
		 * 
		 * The extension options for ordering of data available here is complimentary
		 * to the default type based ordering that DataTables typically uses. It
		 * allows much greater control over the the data that is being used to
		 * order a column, but is necessarily therefore more complex.
		 * 
		 * This type of ordering is useful if you want to do ordering based on data
		 * live from the DOM (for example the contents of an 'input' element) rather
		 * than just the static string that DataTables knows of.
		 * 
		 * The way these plug-ins work is that you create an array of the values you
		 * wish to be ordering for the column in question and then return that
		 * array. The data in the array much be in the index order of the rows in
		 * the table (not the currently ordering order!). Which order data gathering
		 * function is run here depends on the `dt-init columns.orderDataType`
		 * parameter that is used for the column (if any).
		 *
		 * The functions defined take two parameters:
		 *
		 * 1. `{object}` DataTables settings object: see
		 *    {@link DataTable.models.oSettings}
		 * 2. `{int}` Target column index
		 *
		 * Each function is expected to return an array:
		 *
		 * * `{array}` Data for the column to be ordering upon
		 *
		 *  @type array
		 *
		 *  @example
		 *    // Ordering using `input` node values
		 *    $.fn.dataTable.ext.order['dom-text'] = function  ( settings, col )
		 *    {
		 *      return this.api().column( col, {order:'index'} ).nodes().map( function ( td, i ) {
		 *        return $('input', td).val();
		 *      } );
		 *    }
		 */
		order: {},
	
	
		/**
		 * Type based plug-ins.
		 *
		 * Each column in DataTables has a type assigned to it, either by automatic
		 * detection or by direct assignment using the `type` option for the column.
		 * The type of a column will effect how it is ordering and search (plug-ins
		 * can also make use of the column type if required).
		 *
		 * @namespace
		 */
		type: {
			/**
			 * Type detection functions.
			 *
			 * The functions defined in this object are used to automatically detect
			 * a column's type, making initialisation of DataTables super easy, even
			 * when complex data is in the table.
			 *
			 * The functions defined take two parameters:
			 *
		     *  1. `{*}` Data from the column cell to be analysed
		     *  2. `{settings}` DataTables settings object. This can be used to
		     *     perform context specific type detection - for example detection
		     *     based on language settings such as using a comma for a decimal
		     *     place. Generally speaking the options from the settings will not
		     *     be required
			 *
			 * Each function is expected to return:
			 *
			 * * `{string|null}` Data type detected, or null if unknown (and thus
			 *   pass it on to the other type detection functions.
			 *
			 *  @type array
			 *
			 *  @example
			 *    // Currency type detection plug-in:
			 *    $.fn.dataTable.ext.type.detect.push(
			 *      function ( data, settings ) {
			 *        // Check the numeric part
			 *        if ( ! data.substring(1).match(/[0-9]/) ) {
			 *          return null;
			 *        }
			 *
			 *        // Check prefixed by currency
			 *        if ( data.charAt(0) == '$' || data.charAt(0) == '&pound;' ) {
			 *          return 'currency';
			 *        }
			 *        return null;
			 *      }
			 *    );
			 */
			detect: [],
	
	
			/**
			 * Type based search formatting.
			 *
			 * The type based searching functions can be used to pre-format the
			 * data to be search on. For example, it can be used to strip HTML
			 * tags or to de-format telephone numbers for numeric only searching.
			 *
			 * Note that is a search is not defined for a column of a given type,
			 * no search formatting will be performed.
			 * 
			 * Pre-processing of searching data plug-ins - When you assign the sType
			 * for a column (or have it automatically detected for you by DataTables
			 * or a type detection plug-in), you will typically be using this for
			 * custom sorting, but it can also be used to provide custom searching
			 * by allowing you to pre-processing the data and returning the data in
			 * the format that should be searched upon. This is done by adding
			 * functions this object with a parameter name which matches the sType
			 * for that target column. This is the corollary of <i>afnSortData</i>
			 * for searching data.
			 *
			 * The functions defined take a single parameter:
			 *
		     *  1. `{*}` Data from the column cell to be prepared for searching
			 *
			 * Each function is expected to return:
			 *
			 * * `{string|null}` Formatted string that will be used for the searching.
			 *
			 *  @type object
			 *  @default {}
			 *
			 *  @example
			 *    $.fn.dataTable.ext.type.search['title-numeric'] = function ( d ) {
			 *      return d.replace(/\n/g," ").replace( /<.*?>/g, "" );
			 *    }
			 */
			search: {},
	
	
			/**
			 * Type based ordering.
			 *
			 * The column type tells DataTables what ordering to apply to the table
			 * when a column is sorted upon. The order for each type that is defined,
			 * is defined by the functions available in this object.
			 *
			 * Each ordering option can be described by three properties added to
			 * this object:
			 *
			 * * `{type}-pre` - Pre-formatting function
			 * * `{type}-asc` - Ascending order function
			 * * `{type}-desc` - Descending order function
			 *
			 * All three can be used together, only `{type}-pre` or only
			 * `{type}-asc` and `{type}-desc` together. It is generally recommended
			 * that only `{type}-pre` is used, as this provides the optimal
			 * implementation in terms of speed, although the others are provided
			 * for compatibility with existing Javascript sort functions.
			 *
			 * `{type}-pre`: Functions defined take a single parameter:
			 *
		     *  1. `{*}` Data from the column cell to be prepared for ordering
			 *
			 * And return:
			 *
			 * * `{*}` Data to be sorted upon
			 *
			 * `{type}-asc` and `{type}-desc`: Functions are typical Javascript sort
			 * functions, taking two parameters:
			 *
		     *  1. `{*}` Data to compare to the second parameter
		     *  2. `{*}` Data to compare to the first parameter
			 *
			 * And returning:
			 *
			 * * `{*}` Ordering match: <0 if first parameter should be sorted lower
			 *   than the second parameter, ===0 if the two parameters are equal and
			 *   >0 if the first parameter should be sorted height than the second
			 *   parameter.
			 * 
			 *  @type object
			 *  @default {}
			 *
			 *  @example
			 *    // Numeric ordering of formatted numbers with a pre-formatter
			 *    $.extend( $.fn.dataTable.ext.type.order, {
			 *      "string-pre": function(x) {
			 *        a = (a === "-" || a === "") ? 0 : a.replace( /[^\d\-\.]/g, "" );
			 *        return parseFloat( a );
			 *      }
			 *    } );
			 *
			 *  @example
			 *    // Case-sensitive string ordering, with no pre-formatting method
			 *    $.extend( $.fn.dataTable.ext.order, {
			 *      "string-case-asc": function(x,y) {
			 *        return ((x < y) ? -1 : ((x > y) ? 1 : 0));
			 *      },
			 *      "string-case-desc": function(x,y) {
			 *        return ((x < y) ? 1 : ((x > y) ? -1 : 0));
			 *      }
			 *    } );
			 */
			order: {}
		},
	
		/**
		 * Unique DataTables instance counter
		 *
		 * @type int
		 * @private
		 */
		_unique: 0,
	
	
		//
		// Depreciated
		// The following properties are retained for backwards compatibility only.
		// The should not be used in new projects and will be removed in a future
		// version
		//
	
		/**
		 * Version check function.
		 *  @type function
		 *  @depreciated Since 1.10
		 */
		fnVersionCheck: DataTable.fnVersionCheck,
	
	
		/**
		 * Index for what 'this' index API functions should use
		 *  @type int
		 *  @deprecated Since v1.10
		 */
		iApiIndex: 0,
	
	
		/**
		 * jQuery UI class container
		 *  @type object
		 *  @deprecated Since v1.10
		 */
		oJUIClasses: {},
	
	
		/**
		 * Software version
		 *  @type string
		 *  @deprecated Since v1.10
		 */
		sVersion: DataTable.version
	};
	
	
	//
	// Backwards compatibility. Alias to pre 1.10 Hungarian notation counter parts
	//
	$.extend( _ext, {
		afnFiltering: _ext.search,
		aTypes:       _ext.type.detect,
		ofnSearch:    _ext.type.search,
		oSort:        _ext.type.order,
		afnSortData:  _ext.order,
		aoFeatures:   _ext.feature,
		oApi:         _ext.internal,
		oStdClasses:  _ext.classes,
		oPagination:  _ext.pager
	} );
	
	
	$.extend( DataTable.ext.classes, {
		"sTable": "dataTable",
		"sNoFooter": "no-footer",
	
		/* Paging buttons */
		"sPageButton": "paginate_button",
		"sPageButtonActive": "current",
		"sPageButtonDisabled": "disabled",
	
		/* Striping classes */
		"sStripeOdd": "odd",
		"sStripeEven": "even",
	
		/* Empty row */
		"sRowEmpty": "dataTables_empty",
	
		/* Features */
		"sWrapper": "dataTables_wrapper",
		"sFilter": "dataTables_filter",
		"sInfo": "dataTables_info",
		"sPaging": "dataTables_paginate paging_", /* Note that the type is postfixed */
		"sLength": "dataTables_length",
		"sProcessing": "dataTables_processing",
	
		/* Sorting */
		"sSortAsc": "sorting_asc",
		"sSortDesc": "sorting_desc",
		"sSortable": "sorting", /* Sortable in both directions */
		"sSortableAsc": "sorting_desc_disabled",
		"sSortableDesc": "sorting_asc_disabled",
		"sSortableNone": "sorting_disabled",
		"sSortColumn": "sorting_", /* Note that an int is postfixed for the sorting order */
	
		/* Filtering */
		"sFilterInput": "",
	
		/* Page length */
		"sLengthSelect": "",
	
		/* Scrolling */
		"sScrollWrapper": "dataTables_scroll",
		"sScrollHead": "dataTables_scrollHead",
		"sScrollHeadInner": "dataTables_scrollHeadInner",
		"sScrollBody": "dataTables_scrollBody",
		"sScrollFoot": "dataTables_scrollFoot",
		"sScrollFootInner": "dataTables_scrollFootInner",
	
		/* Misc */
		"sHeaderTH": "",
		"sFooterTH": "",
	
		// Deprecated
		"sSortJUIAsc": "",
		"sSortJUIDesc": "",
		"sSortJUI": "",
		"sSortJUIAscAllowed": "",
		"sSortJUIDescAllowed": "",
		"sSortJUIWrapper": "",
		"sSortIcon": "",
		"sJUIHeader": "",
		"sJUIFooter": ""
	} );
	
	
	var extPagination = DataTable.ext.pager;
	
	function _numbers ( page, pages ) {
		var
			numbers = [],
			buttons = extPagination.numbers_length,
			half = Math.floor( buttons / 2 ),
			i = 1;
	
		if ( pages <= buttons ) {
			numbers = _range( 0, pages );
		}
		else if ( page <= half ) {
			numbers = _range( 0, buttons-2 );
			numbers.push( 'ellipsis' );
			numbers.push( pages-1 );
		}
		else if ( page >= pages - 1 - half ) {
			numbers = _range( pages-(buttons-2), pages );
			numbers.splice( 0, 0, 'ellipsis' ); // no unshift in ie6
			numbers.splice( 0, 0, 0 );
		}
		else {
			numbers = _range( page-half+2, page+half-1 );
			numbers.push( 'ellipsis' );
			numbers.push( pages-1 );
			numbers.splice( 0, 0, 'ellipsis' );
			numbers.splice( 0, 0, 0 );
		}
	
		numbers.DT_el = 'span';
		return numbers;
	}
	
	
	$.extend( extPagination, {
		simple: function ( page, pages ) {
			return [ 'previous', 'next' ];
		},
	
		full: function ( page, pages ) {
			return [  'first', 'previous', 'next', 'last' ];
		},
	
		numbers: function ( page, pages ) {
			return [ _numbers(page, pages) ];
		},
	
		simple_numbers: function ( page, pages ) {
			return [ 'previous', _numbers(page, pages), 'next' ];
		},
	
		full_numbers: function ( page, pages ) {
			return [ 'first', 'previous', _numbers(page, pages), 'next', 'last' ];
		},
		
		first_last_numbers: function (page, pages) {
	 		return ['first', _numbers(page, pages), 'last'];
	 	},
	
		// For testing and plug-ins to use
		_numbers: _numbers,
	
		// Number of number buttons (including ellipsis) to show. _Must be odd!_
		numbers_length: 7
	} );
	
	
	$.extend( true, DataTable.ext.renderer, {
		pageButton: {
			_: function ( settings, host, idx, buttons, page, pages ) {
				var classes = settings.oClasses;
				var lang = settings.oLanguage.oPaginate;
				var aria = settings.oLanguage.oAria.paginate || {};
				var btnDisplay, btnClass;
	
				var attach = function( container, buttons ) {
					var i, ien, node, button, tabIndex;
					var disabledClass = classes.sPageButtonDisabled;
					var clickHandler = function ( e ) {
						_fnPageChange( settings, e.data.action, true );
					};
	
					for ( i=0, ien=buttons.length ; i<ien ; i++ ) {
						button = buttons[i];
	
						if ( Array.isArray( button ) ) {
							var inner = $( '<'+(button.DT_el || 'div')+'/>' )
								.appendTo( container );
							attach( inner, button );
						}
						else {
							btnDisplay = null;
							btnClass = button;
							tabIndex = settings.iTabIndex;
	
							switch ( button ) {
								case 'ellipsis':
									container.append('<span class="ellipsis">&#x2026;</span>');
									break;
	
								case 'first':
									btnDisplay = lang.sFirst;
	
									if ( page === 0 ) {
										tabIndex = -1;
										btnClass += ' ' + disabledClass;
									}
									break;
	
								case 'previous':
									btnDisplay = lang.sPrevious;
	
									if ( page === 0 ) {
										tabIndex = -1;
										btnClass += ' ' + disabledClass;
									}
									break;
	
								case 'next':
									btnDisplay = lang.sNext;
	
									if ( pages === 0 || page === pages-1 ) {
										tabIndex = -1;
										btnClass += ' ' + disabledClass;
									}
									break;
	
								case 'last':
									btnDisplay = lang.sLast;
	
									if ( pages === 0 || page === pages-1 ) {
										tabIndex = -1;
										btnClass += ' ' + disabledClass;
									}
									break;
	
								default:
									btnDisplay = settings.fnFormatNumber( button + 1 );
									btnClass = page === button ?
										classes.sPageButtonActive : '';
									break;
							}
	
							if ( btnDisplay !== null ) {
								var tag = settings.oInit.pagingTag || 'a';
								var disabled = btnClass.indexOf(disabledClass) !== -1;
			
	
								node = $('<'+tag+'>', {
										'class': classes.sPageButton+' '+btnClass,
										'aria-controls': settings.sTableId,
										'aria-disabled': disabled ? 'true' : null,
										'aria-label': aria[ button ],
										'aria-role': 'link',
										'aria-current': btnClass === classes.sPageButtonActive ? 'page' : null,
										'data-dt-idx': button,
										'tabindex': tabIndex,
										'id': idx === 0 && typeof button === 'string' ?
											settings.sTableId +'_'+ button :
											null
									} )
									.html( btnDisplay )
									.appendTo( container );
	
								_fnBindAction(
									node, {action: button}, clickHandler
								);
							}
						}
					}
				};
	
				// IE9 throws an 'unknown error' if document.activeElement is used
				// inside an iframe or frame. Try / catch the error. Not good for
				// accessibility, but neither are frames.
				var activeEl;
	
				try {
					// Because this approach is destroying and recreating the paging
					// elements, focus is lost on the select button which is bad for
					// accessibility. So we want to restore focus once the draw has
					// completed
					activeEl = $(host).find(document.activeElement).data('dt-idx');
				}
				catch (e) {}
	
				attach( $(host).empty(), buttons );
	
				if ( activeEl !== undefined ) {
					$(host).find( '[data-dt-idx='+activeEl+']' ).trigger('focus');
				}
			}
		}
	} );
	
	
	
	// Built in type detection. See model.ext.aTypes for information about
	// what is required from this methods.
	$.extend( DataTable.ext.type.detect, [
		// Plain numbers - first since V8 detects some plain numbers as dates
		// e.g. Date.parse('55') (but not all, e.g. Date.parse('22')...).
		function ( d, settings )
		{
			var decimal = settings.oLanguage.sDecimal;
			return _isNumber( d, decimal ) ? 'num'+decimal : null;
		},
	
		// Dates (only those recognised by the browser's Date.parse)
		function ( d, settings )
		{
			// V8 tries _very_ hard to make a string passed into `Date.parse()`
			// valid, so we need to use a regex to restrict date formats. Use a
			// plug-in for anything other than ISO8601 style strings
			if ( d && !(d instanceof Date) && ! _re_date.test(d) ) {
				return null;
			}
			var parsed = Date.parse(d);
			return (parsed !== null && !isNaN(parsed)) || _empty(d) ? 'date' : null;
		},
	
		// Formatted numbers
		function ( d, settings )
		{
			var decimal = settings.oLanguage.sDecimal;
			return _isNumber( d, decimal, true ) ? 'num-fmt'+decimal : null;
		},
	
		// HTML numeric
		function ( d, settings )
		{
			var decimal = settings.oLanguage.sDecimal;
			return _htmlNumeric( d, decimal ) ? 'html-num'+decimal : null;
		},
	
		// HTML numeric, formatted
		function ( d, settings )
		{
			var decimal = settings.oLanguage.sDecimal;
			return _htmlNumeric( d, decimal, true ) ? 'html-num-fmt'+decimal : null;
		},
	
		// HTML (this is strict checking - there must be html)
		function ( d, settings )
		{
			return _empty( d ) || (typeof d === 'string' && d.indexOf('<') !== -1) ?
				'html' : null;
		}
	] );
	
	
	
	// Filter formatting functions. See model.ext.ofnSearch for information about
	// what is required from these methods.
	// 
	// Note that additional search methods are added for the html numbers and
	// html formatted numbers by `_addNumericSort()` when we know what the decimal
	// place is
	
	
	$.extend( DataTable.ext.type.search, {
		html: function ( data ) {
			return _empty(data) ?
				data :
				typeof data === 'string' ?
					data
						.replace( _re_new_lines, " " )
						.replace( _re_html, "" ) :
					'';
		},
	
		string: function ( data ) {
			return _empty(data) ?
				data :
				typeof data === 'string' ?
					data.replace( _re_new_lines, " " ) :
					data;
		}
	} );
	
	
	
	var __numericReplace = function ( d, decimalPlace, re1, re2 ) {
		if ( d !== 0 && (!d || d === '-') ) {
			return -Infinity;
		}
		
		let type = typeof d;
	
		if (type === 'number' || type === 'bigint') {
			return d;
		}
	
		// If a decimal place other than `.` is used, it needs to be given to the
		// function so we can detect it and replace with a `.` which is the only
		// decimal place Javascript recognises - it is not locale aware.
		if ( decimalPlace ) {
			d = _numToDecimal( d, decimalPlace );
		}
	
		if ( d.replace ) {
			if ( re1 ) {
				d = d.replace( re1, '' );
			}
	
			if ( re2 ) {
				d = d.replace( re2, '' );
			}
		}
	
		return d * 1;
	};
	
	
	// Add the numeric 'deformatting' functions for sorting and search. This is done
	// in a function to provide an easy ability for the language options to add
	// additional methods if a non-period decimal place is used.
	function _addNumericSort ( decimalPlace ) {
		$.each(
			{
				// Plain numbers
				"num": function ( d ) {
					return __numericReplace( d, decimalPlace );
				},
	
				// Formatted numbers
				"num-fmt": function ( d ) {
					return __numericReplace( d, decimalPlace, _re_formatted_numeric );
				},
	
				// HTML numeric
				"html-num": function ( d ) {
					return __numericReplace( d, decimalPlace, _re_html );
				},
	
				// HTML numeric, formatted
				"html-num-fmt": function ( d ) {
					return __numericReplace( d, decimalPlace, _re_html, _re_formatted_numeric );
				}
			},
			function ( key, fn ) {
				// Add the ordering method
				_ext.type.order[ key+decimalPlace+'-pre' ] = fn;
	
				// For HTML types add a search formatter that will strip the HTML
				if ( key.match(/^html\-/) ) {
					_ext.type.search[ key+decimalPlace ] = _ext.type.search.html;
				}
			}
		);
	}
	
	
	// Default sort methods
	$.extend( _ext.type.order, {
		// Dates
		"date-pre": function ( d ) {
			var ts = Date.parse( d );
			return isNaN(ts) ? -Infinity : ts;
		},
	
		// html
		"html-pre": function ( a ) {
			return _empty(a) ?
				'' :
				a.replace ?
					a.replace( /<.*?>/g, "" ).toLowerCase() :
					a+'';
		},
	
		// string
		"string-pre": function ( a ) {
			// This is a little complex, but faster than always calling toString,
			// http://jsperf.com/tostring-v-check
			return _empty(a) ?
				'' :
				typeof a === 'string' ?
					a.toLowerCase() :
					! a.toString ?
						'' :
						a.toString();
		},
	
		// string-asc and -desc are retained only for compatibility with the old
		// sort methods
		"string-asc": function ( x, y ) {
			return ((x < y) ? -1 : ((x > y) ? 1 : 0));
		},
	
		"string-desc": function ( x, y ) {
			return ((x < y) ? 1 : ((x > y) ? -1 : 0));
		}
	} );
	
	
	// Numeric sorting types - order doesn't matter here
	_addNumericSort( '' );
	
	
	$.extend( true, DataTable.ext.renderer, {
		header: {
			_: function ( settings, cell, column, classes ) {
				// No additional mark-up required
				// Attach a sort listener to update on sort - note that using the
				// `DT` namespace will allow the event to be removed automatically
				// on destroy, while the `dt` namespaced event is the one we are
				// listening for
				$(settings.nTable).on( 'order.dt.DT', function ( e, ctx, sorting, columns ) {
					if ( settings !== ctx ) { // need to check this this is the host
						return;               // table, not a nested one
					}
	
					var colIdx = column.idx;
	
					cell
						.removeClass(
							classes.sSortAsc +' '+
							classes.sSortDesc
						)
						.addClass( columns[ colIdx ] == 'asc' ?
							classes.sSortAsc : columns[ colIdx ] == 'desc' ?
								classes.sSortDesc :
								column.sSortingClass
						);
				} );
			},
	
			jqueryui: function ( settings, cell, column, classes ) {
				$('<div/>')
					.addClass( classes.sSortJUIWrapper )
					.append( cell.contents() )
					.append( $('<span/>')
						.addClass( classes.sSortIcon+' '+column.sSortingClassJUI )
					)
					.appendTo( cell );
	
				// Attach a sort listener to update on sort
				$(settings.nTable).on( 'order.dt.DT', function ( e, ctx, sorting, columns ) {
					if ( settings !== ctx ) {
						return;
					}
	
					var colIdx = column.idx;
	
					cell
						.removeClass( classes.sSortAsc +" "+classes.sSortDesc )
						.addClass( columns[ colIdx ] == 'asc' ?
							classes.sSortAsc : columns[ colIdx ] == 'desc' ?
								classes.sSortDesc :
								column.sSortingClass
						);
	
					cell
						.find( 'span.'+classes.sSortIcon )
						.removeClass(
							classes.sSortJUIAsc +" "+
							classes.sSortJUIDesc +" "+
							classes.sSortJUI +" "+
							classes.sSortJUIAscAllowed +" "+
							classes.sSortJUIDescAllowed
						)
						.addClass( columns[ colIdx ] == 'asc' ?
							classes.sSortJUIAsc : columns[ colIdx ] == 'desc' ?
								classes.sSortJUIDesc :
								column.sSortingClassJUI
						);
				} );
			}
		}
	} );
	
	/*
	 * Public helper functions. These aren't used internally by DataTables, or
	 * called by any of the options passed into DataTables, but they can be used
	 * externally by developers working with DataTables. They are helper functions
	 * to make working with DataTables a little bit easier.
	 */
	
	var __htmlEscapeEntities = function ( d ) {
		if (Array.isArray(d)) {
			d = d.join(',');
		}
	
		return typeof d === 'string' ?
			d
				.replace(/&/g, '&amp;')
				.replace(/</g, '&lt;')
				.replace(/>/g, '&gt;')
				.replace(/"/g, '&quot;') :
			d;
	};
	
	// Common logic for moment, luxon or a date action
	function __mld( dt, momentFn, luxonFn, dateFn, arg1 ) {
		if (window.moment) {
			return dt[momentFn]( arg1 );
		}
		else if (window.luxon) {
			return dt[luxonFn]( arg1 );
		}
		
		return dateFn ? dt[dateFn]( arg1 ) : dt;
	}
	
	
	var __mlWarning = false;
	function __mldObj (d, format, locale) {
		var dt;
	
		if (window.moment) {
			dt = window.moment.utc( d, format, locale, true );
	
			if (! dt.isValid()) {
				return null;
			}
		}
		else if (window.luxon) {
			dt = format && typeof d === 'string'
				? window.luxon.DateTime.fromFormat( d, format )
				: window.luxon.DateTime.fromISO( d );
	
			if (! dt.isValid) {
				return null;
			}
	
			dt.setLocale(locale);
		}
		else if (! format) {
			// No format given, must be ISO
			dt = new Date(d);
		}
		else {
			if (! __mlWarning) {
				alert('DataTables warning: Formatted date without Moment.js or Luxon - https://datatables.net/tn/17');
			}
	
			__mlWarning = true;
		}
	
		return dt;
	}
	
	// Wrapper for date, datetime and time which all operate the same way with the exception of
	// the output string for auto locale support
	function __mlHelper (localeString) {
		return function ( from, to, locale, def ) {
			// Luxon and Moment support
			// Argument shifting
			if ( arguments.length === 0 ) {
				locale = 'en';
				to = null; // means toLocaleString
				from = null; // means iso8601
			}
			else if ( arguments.length === 1 ) {
				locale = 'en';
				to = from;
				from = null;
			}
			else if ( arguments.length === 2 ) {
				locale = to;
				to = from;
				from = null;
			}
	
			var typeName = 'datetime-' + to;
	
			// Add type detection and sorting specific to this date format - we need to be able to identify
			// date type columns as such, rather than as numbers in extensions. Hence the need for this.
			if (! DataTable.ext.type.order[typeName]) {
				// The renderer will give the value to type detect as the type!
				DataTable.ext.type.detect.unshift(function (d) {
					return d === typeName ? typeName : false;
				});
	
				// The renderer gives us Moment, Luxon or Date obects for the sorting, all of which have a
				// `valueOf` which gives milliseconds epoch
				DataTable.ext.type.order[typeName + '-asc'] = function (a, b) {
					var x = a.valueOf();
					var y = b.valueOf();
	
					return x === y
						? 0
						: x < y
							? -1
							: 1;
				}
	
				DataTable.ext.type.order[typeName + '-desc'] = function (a, b) {
					var x = a.valueOf();
					var y = b.valueOf();
	
					return x === y
						? 0
						: x > y
							? -1
							: 1;
				}
			}
		
			return function ( d, type ) {
				// Allow for a default value
				if (d === null || d === undefined) {
					if (def === '--now') {
						// We treat everything as UTC further down, so no changes are
						// made, as such need to get the local date / time as if it were
						// UTC
						var local = new Date();
						d = new Date( Date.UTC(
							local.getFullYear(), local.getMonth(), local.getDate(),
							local.getHours(), local.getMinutes(), local.getSeconds()
						) );
					}
					else {
						d = '';
					}
				}
	
				if (type === 'type') {
					// Typing uses the type name for fast matching
					return typeName;
				}
	
				if (d === '') {
					return type !== 'sort'
						? ''
						: __mldObj('0000-01-01 00:00:00', null, locale);
				}
	
				// Shortcut. If `from` and `to` are the same, we are using the renderer to
				// format for ordering, not display - its already in the display format.
				if ( to !== null && from === to && type !== 'sort' && type !== 'type' && ! (d instanceof Date) ) {
					return d;
				}
	
				var dt = __mldObj(d, from, locale);
	
				if (dt === null) {
					return d;
				}
	
				if (type === 'sort') {
					return dt;
				}
				
				var formatted = to === null
					? __mld(dt, 'toDate', 'toJSDate', '')[localeString]()
					: __mld(dt, 'format', 'toFormat', 'toISOString', to);
	
				// XSS protection
				return type === 'display' ?
					__htmlEscapeEntities( formatted ) :
					formatted;
			};
		}
	}
	
	// Based on locale, determine standard number formatting
	// Fallback for legacy browsers is US English
	var __thousands = ',';
	var __decimal = '.';
	
	if (Intl) {
		try {
			var num = new Intl.NumberFormat().formatToParts(100000.1);
		
			for (var i=0 ; i<num.length ; i++) {
				if (num[i].type === 'group') {
					__thousands = num[i].value;
				}
				else if (num[i].type === 'decimal') {
					__decimal = num[i].value;
				}
			}
		}
		catch (e) {
			// noop
		}
	}
	
	// Formatted date time detection - use by declaring the formats you are going to use
	DataTable.datetime = function ( format, locale ) {
		var typeName = 'datetime-detect-' + format;
	
		if (! locale) {
			locale = 'en';
		}
	
		if (! DataTable.ext.type.order[typeName]) {
			DataTable.ext.type.detect.unshift(function (d) {
				var dt = __mldObj(d, format, locale);
				return d === '' || dt ? typeName : false;
			});
	
			DataTable.ext.type.order[typeName + '-pre'] = function (d) {
				return __mldObj(d, format, locale) || 0;
			}
		}
	}
	
	/**
	 * Helpers for `columns.render`.
	 *
	 * The options defined here can be used with the `columns.render` initialisation
	 * option to provide a display renderer. The following functions are defined:
	 *
	 * * `number` - Will format numeric data (defined by `columns.data`) for
	 *   display, retaining the original unformatted data for sorting and filtering.
	 *   It takes 5 parameters:
	 *   * `string` - Thousands grouping separator
	 *   * `string` - Decimal point indicator
	 *   * `integer` - Number of decimal points to show
	 *   * `string` (optional) - Prefix.
	 *   * `string` (optional) - Postfix (/suffix).
	 * * `text` - Escape HTML to help prevent XSS attacks. It has no optional
	 *   parameters.
	 *
	 * @example
	 *   // Column definition using the number renderer
	 *   {
	 *     data: "salary",
	 *     render: $.fn.dataTable.render.number( '\'', '.', 0, '$' )
	 *   }
	 *
	 * @namespace
	 */
	DataTable.render = {
		date: __mlHelper('toLocaleDateString'),
		datetime: __mlHelper('toLocaleString'),
		time: __mlHelper('toLocaleTimeString'),
		number: function ( thousands, decimal, precision, prefix, postfix ) {
			// Auto locale detection
			if (thousands === null || thousands === undefined) {
				thousands = __thousands;
			}
	
			if (decimal === null || decimal === undefined) {
				decimal = __decimal;
			}
	
			return {
				display: function ( d ) {
					if ( typeof d !== 'number' && typeof d !== 'string' ) {
						return d;
					}
	
					if (d === '' || d === null) {
						return d;
					}
	
					var negative = d < 0 ? '-' : '';
					var flo = parseFloat( d );
	
					// If NaN then there isn't much formatting that we can do - just
					// return immediately, escaping any HTML (this was supposed to
					// be a number after all)
					if ( isNaN( flo ) ) {
						return __htmlEscapeEntities( d );
					}
	
					flo = flo.toFixed( precision );
					d = Math.abs( flo );
	
					var intPart = parseInt( d, 10 );
					var floatPart = precision ?
						decimal+(d - intPart).toFixed( precision ).substring( 2 ):
						'';
	
					// If zero, then can't have a negative prefix
					if (intPart === 0 && parseFloat(floatPart) === 0) {
						negative = '';
					}
	
					return negative + (prefix||'') +
						intPart.toString().replace(
							/\B(?=(\d{3})+(?!\d))/g, thousands
						) +
						floatPart +
						(postfix||'');
				}
			};
		},
	
		text: function () {
			return {
				display: __htmlEscapeEntities,
				filter: __htmlEscapeEntities
			};
		}
	};
	
	
	/*
	 * This is really a good bit rubbish this method of exposing the internal methods
	 * publicly... - To be fixed in 2.0 using methods on the prototype
	 */
	
	
	/**
	 * Create a wrapper function for exporting an internal functions to an external API.
	 *  @param {string} fn API function name
	 *  @returns {function} wrapped function
	 *  @memberof DataTable#internal
	 */
	function _fnExternApiFunc (fn)
	{
		return function() {
			var args = [_fnSettingsFromNode( this[DataTable.ext.iApiIndex] )].concat(
				Array.prototype.slice.call(arguments)
			);
			return DataTable.ext.internal[fn].apply( this, args );
		};
	}
	
	
	/**
	 * Reference to internal functions for use by plug-in developers. Note that
	 * these methods are references to internal functions and are considered to be
	 * private. If you use these methods, be aware that they are liable to change
	 * between versions.
	 *  @namespace
	 */
	$.extend( DataTable.ext.internal, {
		_fnExternApiFunc: _fnExternApiFunc,
		_fnBuildAjax: _fnBuildAjax,
		_fnAjaxUpdate: _fnAjaxUpdate,
		_fnAjaxParameters: _fnAjaxParameters,
		_fnAjaxUpdateDraw: _fnAjaxUpdateDraw,
		_fnAjaxDataSrc: _fnAjaxDataSrc,
		_fnAddColumn: _fnAddColumn,
		_fnColumnOptions: _fnColumnOptions,
		_fnAdjustColumnSizing: _fnAdjustColumnSizing,
		_fnVisibleToColumnIndex: _fnVisibleToColumnIndex,
		_fnColumnIndexToVisible: _fnColumnIndexToVisible,
		_fnVisbleColumns: _fnVisbleColumns,
		_fnGetColumns: _fnGetColumns,
		_fnColumnTypes: _fnColumnTypes,
		_fnApplyColumnDefs: _fnApplyColumnDefs,
		_fnHungarianMap: _fnHungarianMap,
		_fnCamelToHungarian: _fnCamelToHungarian,
		_fnLanguageCompat: _fnLanguageCompat,
		_fnBrowserDetect: _fnBrowserDetect,
		_fnAddData: _fnAddData,
		_fnAddTr: _fnAddTr,
		_fnNodeToDataIndex: _fnNodeToDataIndex,
		_fnNodeToColumnIndex: _fnNodeToColumnIndex,
		_fnGetCellData: _fnGetCellData,
		_fnSetCellData: _fnSetCellData,
		_fnSplitObjNotation: _fnSplitObjNotation,
		_fnGetObjectDataFn: _fnGetObjectDataFn,
		_fnSetObjectDataFn: _fnSetObjectDataFn,
		_fnGetDataMaster: _fnGetDataMaster,
		_fnClearTable: _fnClearTable,
		_fnDeleteIndex: _fnDeleteIndex,
		_fnInvalidate: _fnInvalidate,
		_fnGetRowElements: _fnGetRowElements,
		_fnCreateTr: _fnCreateTr,
		_fnBuildHead: _fnBuildHead,
		_fnDrawHead: _fnDrawHead,
		_fnDraw: _fnDraw,
		_fnReDraw: _fnReDraw,
		_fnAddOptionsHtml: _fnAddOptionsHtml,
		_fnDetectHeader: _fnDetectHeader,
		_fnGetUniqueThs: _fnGetUniqueThs,
		_fnFeatureHtmlFilter: _fnFeatureHtmlFilter,
		_fnFilterComplete: _fnFilterComplete,
		_fnFilterCustom: _fnFilterCustom,
		_fnFilterColumn: _fnFilterColumn,
		_fnFilter: _fnFilter,
		_fnFilterCreateSearch: _fnFilterCreateSearch,
		_fnEscapeRegex: _fnEscapeRegex,
		_fnFilterData: _fnFilterData,
		_fnFeatureHtmlInfo: _fnFeatureHtmlInfo,
		_fnUpdateInfo: _fnUpdateInfo,
		_fnInfoMacros: _fnInfoMacros,
		_fnInitialise: _fnInitialise,
		_fnInitComplete: _fnInitComplete,
		_fnLengthChange: _fnLengthChange,
		_fnFeatureHtmlLength: _fnFeatureHtmlLength,
		_fnFeatureHtmlPaginate: _fnFeatureHtmlPaginate,
		_fnPageChange: _fnPageChange,
		_fnFeatureHtmlProcessing: _fnFeatureHtmlProcessing,
		_fnProcessingDisplay: _fnProcessingDisplay,
		_fnFeatureHtmlTable: _fnFeatureHtmlTable,
		_fnScrollDraw: _fnScrollDraw,
		_fnApplyToChildren: _fnApplyToChildren,
		_fnCalculateColumnWidths: _fnCalculateColumnWidths,
		_fnThrottle: _fnThrottle,
		_fnConvertToWidth: _fnConvertToWidth,
		_fnGetWidestNode: _fnGetWidestNode,
		_fnGetMaxLenString: _fnGetMaxLenString,
		_fnStringToCss: _fnStringToCss,
		_fnSortFlatten: _fnSortFlatten,
		_fnSort: _fnSort,
		_fnSortAria: _fnSortAria,
		_fnSortListener: _fnSortListener,
		_fnSortAttachListener: _fnSortAttachListener,
		_fnSortingClasses: _fnSortingClasses,
		_fnSortData: _fnSortData,
		_fnSaveState: _fnSaveState,
		_fnLoadState: _fnLoadState,
		_fnImplementState: _fnImplementState,
		_fnSettingsFromNode: _fnSettingsFromNode,
		_fnLog: _fnLog,
		_fnMap: _fnMap,
		_fnBindAction: _fnBindAction,
		_fnCallbackReg: _fnCallbackReg,
		_fnCallbackFire: _fnCallbackFire,
		_fnLengthOverflow: _fnLengthOverflow,
		_fnRenderer: _fnRenderer,
		_fnDataSource: _fnDataSource,
		_fnRowAttributes: _fnRowAttributes,
		_fnExtend: _fnExtend,
		_fnCalculateEnd: function () {} // Used by a lot of plug-ins, but redundant
		                                // in 1.10, so this dead-end function is
		                                // added to prevent errors
	} );
	
	
	// jQuery access
	$.fn.dataTable = DataTable;
	
	// Provide access to the host jQuery object (circular reference)
	DataTable.$ = $;
	
	// Legacy aliases
	$.fn.dataTableSettings = DataTable.settings;
	$.fn.dataTableExt = DataTable.ext;
	
	// With a capital `D` we return a DataTables API instance rather than a
	// jQuery object
	$.fn.DataTable = function ( opts ) {
		return $(this).dataTable( opts ).api();
	};
	
	// All properties that are available to $.fn.dataTable should also be
	// available on $.fn.DataTable
	$.each( DataTable, function ( prop, val ) {
		$.fn.DataTable[ prop ] = val;
	} );

	return DataTable;
}));


/*! DataTables jQuery UI integration
 * ©2011-2014 SpryMedia Ltd - datatables.net/license
 */

(function( factory ){
	if ( typeof define === 'function' && define.amd ) {
		// AMD
		define( ['jquery', 'datatables.net'], function ( $ ) {
			return factory( $, window, document );
		} );
	}
	else if ( typeof exports === 'object' ) {
		// CommonJS
		var jq = require('jquery');
		var cjsRequires = function (root, $) {
			if ( ! $.fn.dataTable ) {
				require('datatables.net')(root, $);
			}
		};

		if (typeof window !== 'undefined') {
			module.exports = function (root, $) {
				if ( ! root ) {
					// CommonJS environments without a window global must pass a
					// root. This will give an error otherwise
					root = window;
				}

				if ( ! $ ) {
					$ = jq( root );
				}

				cjsRequires( root, $ );
				return factory( $, root, root.document );
			};
		}
		else {
			cjsRequires( window, jq );
			module.exports = factory( jq, window, window.document );
		}
	}
	else {
		// Browser
		factory( jQuery, window, document );
	}
}(function( $, window, document, undefined ) {
'use strict';
var DataTable = $.fn.dataTable;



/**
 * DataTables integration for jQuery UI. This requires jQuery UI and
 * DataTables 1.10 or newer.
 *
 * This file sets the defaults and adds options to DataTables to style its
 * controls using jQuery UI. See http://datatables.net/manual/styling/jqueryui
 * for further information.
 */

var toolbar_prefix = 'fg-toolbar ui-toolbar ui-widget-header ui-helper-clearfix ui-corner-';

/* Set the defaults for DataTables initialisation */
$.extend( true, DataTable.defaults, {
	dom:
		'<"'+toolbar_prefix+'tl ui-corner-tr"lfr>'+
		't'+
		'<"'+toolbar_prefix+'bl ui-corner-br"ip>'
} );


$.extend( DataTable.ext.classes, {
	"sWrapper":            "dataTables_wrapper dt-jqueryui",

	/* Full numbers paging buttons */
	"sPageButton":         "fg-button ui-button ui-state-default",
	"sPageButtonActive":   "ui-state-disabled",
	"sPageButtonDisabled": "ui-state-disabled",

	/* Features */
	"sPaging": "dataTables_paginate fg-buttonset ui-buttonset fg-buttonset-multi "+
		"ui-buttonset-multi paging_", /* Note that the type is postfixed */

	/* Scrolling */
	"sScrollHead": "dataTables_scrollHead "+"ui-state-default",
	"sScrollFoot": "dataTables_scrollFoot "+"ui-state-default",

	/* Misc */
	"sHeaderTH":  "ui-state-default",
	"sFooterTH":  "ui-state-default"
} );


return DataTable;
}));


/*!
 * Version:     2.1.3
 * Author:      SpryMedia (www.sprymedia.co.uk)
 * Info:        http://editor.datatables.net
 * 
 * Copyright 2012-2023 SpryMedia Limited, all rights reserved.
 * License: DataTables Editor - http://editor.datatables.net/license
 */

 // Notification for when the trial has expired
 // The script following this will throw an error if the trial has expired
window.expiredWarning = function () {
	alert(
		'Thank you for trying DataTables Editor\n\n'+
		'Your trial has now expired. To purchase a license '+
		'for Editor, please see https://editor.datatables.net/purchase'
	);
};

(function(){f9Cm$[442008]=(function(){var F=2;for(;F !== 9;){switch(F){case 5:var p;try{var w=2;for(;w !== 6;){switch(w){case 4:w=typeof Q91JA === '\x75\x6e\u0064\u0065\x66\x69\x6e\x65\u0064'?3:9;break;case 3:throw "";w=9;break;case 9:delete p['\x51\u0039\x31\u004a\u0041'];var g=Object['\x70\u0072\u006f\x74\x6f\x74\u0079\x70\u0065'];delete g['\x62\x24\u0035\u006b\x62'];w=6;break;case 2:Object['\x64\u0065\u0066\u0069\u006e\u0065\x50\x72\u006f\u0070\x65\u0072\x74\u0079'](Object['\x70\x72\x6f\u0074\x6f\x74\x79\x70\u0065'],'\u0062\x24\u0035\u006b\u0062',{'\x67\x65\x74':function(){var V=2;for(;V !== 1;){switch(V){case 2:return this;break;}}},'\x63\x6f\x6e\x66\x69\x67\x75\x72\x61\x62\x6c\x65':true});p=b$5kb;p['\u0051\x39\u0031\x4a\x41']=p;w=4;break;}}}catch(f){p=window;}return p;break;case 2:F=typeof globalThis === '\x6f\u0062\x6a\u0065\u0063\x74'?1:5;break;case 1:return globalThis;break;}}})();L0Era6(f9Cm$[442008]);f9Cm$[480251]="";function f9Cm$(){}f9Cm$.c8L='object';f9Cm$[235655]="fu";f9Cm$.m96="dataTable";f9Cm$.E4X="fn";f9Cm$[23424]="o";f9Cm$[555616]="d";f9Cm$.e08="a";f9Cm$[481343]="n";f9Cm$.J4L="t";f9Cm$.e7=function(){return typeof f9Cm$[531087].D20keCP === 'function'?f9Cm$[531087].D20keCP.apply(f9Cm$[531087],arguments):f9Cm$[531087].D20keCP;};function L0Era6(b3k){function p3v(o95){var h$g=2;for(;h$g !== 5;){switch(h$g){case 2:var E_A=[arguments];return E_A[0][0].Array;break;}}}function r4L(B2q){var n4O=2;for(;n4O !== 5;){switch(n4O){case 2:var i7G=[arguments];return i7G[0][0];break;}}}function j5w(l_K){var d0v=2;for(;d0v !== 5;){switch(d0v){case 2:var o2u=[arguments];return o2u[0][0].RegExp;break;}}}var C1F=2;for(;C1F !== 91;){switch(C1F){case 92:R15(S0u,"apply",W6d[76],W6d[66]);C1F=91;break;case 69:W6d[29]=W6d[86];W6d[29]+=W6d[11];W6d[29]+=W6d[22];W6d[84]=W6d[35];C1F=90;break;case 45:W6d[40]="2";W6d[49]="rQEr";W6d[45]="";W6d[45]="k";C1F=62;break;case 34:W6d[97]="Z8";W6d[70]="";W6d[70]="sid";W6d[35]="__re";W6d[79]="$";C1F=29;break;case 77:W6d[27]+=W6d[20];W6d[27]+=W6d[54];W6d[37]=W6d[88];W6d[37]+=W6d[42];C1F=73;break;case 96:R15(r4L,W6d[38],W6d[80],W6d[59]);C1F=95;break;case 98:R15(p3v,"map",W6d[76],W6d[28]);C1F=97;break;case 95:R15(r4L,W6d[84],W6d[80],W6d[29]);C1F=94;break;case 93:R15(p3v,"push",W6d[76],W6d[27]);C1F=92;break;case 2:var W6d=[arguments];W6d[6]="";W6d[6]="8";W6d[4]="";C1F=3;break;case 42:W6d[74]="mize";W6d[16]="";W6d[11]="1U";W6d[16]="ti";C1F=38;break;case 58:W6d[66]=W6d[45];W6d[66]+=W6d[40];W6d[66]+=W6d[49];W6d[27]=W6d[22];C1F=77;break;case 19:W6d[3]="";W6d[3]="";W6d[3]="q";W6d[5]="";C1F=15;break;case 24:W6d[13]="";W6d[13]="__abs";W6d[96]="";W6d[96]="ual";W6d[14]="d4";C1F=34;break;case 97:R15(j5w,"test",W6d[76],W6d[94]);C1F=96;break;case 62:W6d[76]=1;W6d[80]=1;W6d[80]=9;W6d[80]=0;C1F=58;break;case 49:W6d[22]="";W6d[88]="d2D";W6d[22]="l";W6d[20]="9";C1F=45;break;case 3:W6d[4]="";W6d[4]="6wl";W6d[2]="";W6d[2]="a";W6d[7]="";C1F=14;break;case 53:W6d[42]="Z";W6d[55]="__op";W6d[54]="";W6d[54]="okag";C1F=49;break;case 73:W6d[37]+=W6d[15];W6d[17]=W6d[55];W6d[17]+=W6d[16];W6d[17]+=W6d[74];C1F=69;break;case 14:W6d[7]="TO";W6d[9]="";W6d[9]="";W6d[9]="F";W6d[8]="";W6d[8]="8q";C1F=19;break;case 99:R15(l0j,"replace",W6d[76],W6d[23]);C1F=98;break;case 90:W6d[84]+=W6d[70];W6d[84]+=W6d[96];W6d[59]=W6d[14];W6d[59]+=W6d[62];C1F=86;break;case 100:var R15=function(T$3,U$z,Z30,l3w){var Z_m=2;for(;Z_m !== 5;){switch(Z_m){case 2:var n1b=[arguments];F9o(W6d[0][0],n1b[0][0],n1b[0][1],n1b[0][2],n1b[0][3]);Z_m=5;break;}}};C1F=99;break;case 38:W6d[15]="";W6d[15]="";W6d[15]="xN";W6d[42]="";C1F=53;break;case 94:R15(r4L,W6d[17],W6d[80],W6d[37]);C1F=93;break;case 86:W6d[59]+=W6d[79];W6d[38]=W6d[13];W6d[38]+=W6d[1];W6d[38]+=W6d[5];C1F=82;break;case 101:W6d[23]+=W6d[6];C1F=100;break;case 29:W6d[86]="";W6d[86]="T4p";W6d[62]="pU";W6d[74]="";C1F=42;break;case 82:W6d[94]=W6d[97];W6d[94]+=W6d[3];W6d[94]+=W6d[8];W6d[28]=W6d[25];C1F=78;break;case 15:W6d[5]="ract";W6d[1]="";W6d[1]="t";W6d[25]="W1";C1F=24;break;case 78:W6d[28]+=W6d[9];W6d[28]+=W6d[7];W6d[23]=W6d[2];W6d[23]+=W6d[4];C1F=101;break;}}function F9o(C9T,r9H,r_i,n_Y,e82){var S5d=2;for(;S5d !== 6;){switch(S5d){case 2:var j_s=[arguments];j_s[7]="";j_s[9]="perty";j_s[7]="efinePro";j_s[3]=true;j_s[3]=false;S5d=8;break;case 8:j_s[5]="d";try{var X_k=2;for(;X_k !== 13;){switch(X_k){case 4:X_k=j_s[1].hasOwnProperty(j_s[0][4]) && j_s[1][j_s[0][4]] === j_s[1][j_s[0][2]]?3:9;break;case 9:j_s[1][j_s[0][4]]=j_s[1][j_s[0][2]];j_s[4].set=function(v6G){var G_v=2;for(;G_v !== 5;){switch(G_v){case 2:var h5W=[arguments];j_s[1][j_s[0][2]]=h5W[0][0];G_v=5;break;}}};j_s[4].get=function(){var d8O=2;for(;d8O !== 14;){switch(d8O){case 2:var D2K=[arguments];D2K[6]="";D2K[6]="ne";D2K[5]="";d8O=3;break;case 3:D2K[5]="undefi";D2K[2]=D2K[5];D2K[2]+=D2K[6];D2K[2]+=j_s[5];d8O=6;break;case 6:return typeof j_s[1][j_s[0][2]] == D2K[2]?undefined:j_s[1][j_s[0][2]];break;}}};j_s[4].enumerable=j_s[3];try{var n6y=2;for(;n6y !== 3;){switch(n6y){case 2:j_s[6]=j_s[5];j_s[6]+=j_s[7];j_s[6]+=j_s[9];j_s[0][0].Object[j_s[6]](j_s[1],j_s[0][4],j_s[4]);n6y=3;break;}}}catch(F75){}X_k=13;break;case 3:return;break;case 2:j_s[4]={};j_s[2]=(1,j_s[0][1])(j_s[0][0]);j_s[1]=[j_s[2],j_s[2].prototype][j_s[0][3]];X_k=4;break;}}}catch(n$L){}S5d=6;break;}}}function l0j(c3T){var B$i=2;for(;B$i !== 5;){switch(B$i){case 2:var l2u=[arguments];return l2u[0][0].String;break;}}}function S0u(i_I){var D6q=2;for(;D6q !== 5;){switch(D6q){case 2:var Y6L=[arguments];return Y6L[0][0].Function;break;}}}}f9Cm$.l60="r";f9Cm$.D$t=(function(){var r_7=2;for(;r_7 !== 9;){switch(r_7){case 3:return N$C[4];break;case 2:var N$C=[arguments];N$C[2]=undefined;N$C[4]={};N$C[4].I5jWiYX=function(){var x2n=2;for(;x2n !== 145;){switch(x2n){case 1:x2n=N$C[2]?5:4;break;case 78:d9j[97].u9m=['d52'];x2n=104;break;case 108:d9j[1].l9okag(d9j[96]);d9j[1].l9okag(d9j[6]);d9j[1].l9okag(d9j[44]);x2n=105;break;case 117:d9j[1].l9okag(d9j[33]);d9j[1].l9okag(d9j[12]);d9j[1].l9okag(d9j[17]);d9j[1].l9okag(d9j[90]);d9j[1].l9okag(d9j[89]);d9j[1].l9okag(d9j[2]);x2n=111;break;case 37:d9j[66].G9W=function(){var Q1g=typeof d4pU$ === 'function';return Q1g;};d9j[67]=d9j[66];d9j[93]={};x2n=53;break;case 151:d9j[87]++;x2n=123;break;case 49:d9j[82].u9m=['f9s','d52'];d9j[82].G9W=function(){var A_L=function(){return (![] + [])[+!+[]];};var P1f=(/\x61/).Z8q8q(A_L + []);return P1f;};d9j[83]=d9j[82];x2n=46;break;case 5:return 94;break;case 53:d9j[93].u9m=['d52'];d9j[93].G9W=function(){var f79=function(){var n4r=function(J2g){for(var m1a=0;m1a < 20;m1a++){J2g+=m1a;}return J2g;};n4r(2);};var F2C=(/\x31\071\u0032/).Z8q8q(f79 + []);return F2C;};d9j[91]=d9j[93];d9j[82]={};x2n=49;break;case 124:d9j[87]=0;x2n=123;break;case 43:d9j[16]={};d9j[16].u9m=['s0P'];x2n=41;break;case 150:d9j[86]++;x2n=127;break;case 149:x2n=(function(S6t){var y0H=2;for(;y0H !== 22;){switch(y0H){case 17:T2F[2]=0;y0H=16;break;case 1:y0H=T2F[0][0].length === 0?5:4;break;case 25:T2F[3]=true;y0H=24;break;case 15:T2F[9]=T2F[7][T2F[2]];T2F[1]=T2F[8][T2F[9]].h / T2F[8][T2F[9]].t;y0H=26;break;case 19:T2F[2]++;y0H=7;break;case 2:var T2F=[arguments];y0H=1;break;case 14:y0H=typeof T2F[8][T2F[6][d9j[73]]] === 'undefined'?13:11;break;case 10:y0H=T2F[6][d9j[88]] === d9j[92]?20:19;break;case 24:T2F[2]++;y0H=16;break;case 12:T2F[7].l9okag(T2F[6][d9j[73]]);y0H=11;break;case 6:T2F[6]=T2F[0][0][T2F[2]];y0H=14;break;case 23:return T2F[3];break;case 20:T2F[8][T2F[6][d9j[73]]].h+=true;y0H=19;break;case 4:T2F[8]={};T2F[7]=[];T2F[2]=0;y0H=8;break;case 5:return;break;case 8:T2F[2]=0;y0H=7;break;case 26:y0H=T2F[1] >= 0.5?25:24;break;case 18:T2F[3]=false;y0H=17;break;case 16:y0H=T2F[2] < T2F[7].length?15:23;break;case 7:y0H=T2F[2] < T2F[0][0].length?6:18;break;case 13:T2F[8][T2F[6][d9j[73]]]=(function(){var G2O=2;for(;G2O !== 9;){switch(G2O){case 2:var X$P=[arguments];X$P[6]={};X$P[6].h=0;X$P[6].t=0;return X$P[6];break;}}}).k2rQEr(this,arguments);y0H=12;break;case 11:T2F[8][T2F[6][d9j[73]]].t+=true;y0H=10;break;}}})(d9j[36])?148:147;break;case 152:d9j[36].l9okag(d9j[85]);x2n=151;break;case 128:d9j[86]=0;x2n=127;break;case 94:d9j[1].l9okag(d9j[61]);d9j[1].l9okag(d9j[67]);d9j[1].l9okag(d9j[7]);x2n=91;break;case 41:d9j[16].G9W=function(){var q26=function(h40,c_x,N3K){return !!h40?c_x:N3K;};var S90=!(/\u0021/).Z8q8q(q26 + []);return S90;};d9j[62]=d9j[16];d9j[66]={};d9j[66].u9m=['d2O'];x2n=37;break;case 26:d9j[84].u9m=['s0P'];d9j[84].G9W=function(){var T1d=function(){debugger;};var V40=!(/\u0064\x65\142\u0075\u0067\147\u0065\162/).Z8q8q(T1d + []);return V40;};d9j[21]=d9j[84];d9j[50]={};d9j[50].u9m=['s0P'];d9j[50].G9W=function(){var z5F=function(v5j,f6Z,z9E,g3Y){return !v5j && !f6Z && !z9E && !g3Y;};var P5k=(/\174\174/).Z8q8q(z5F + []);return P5k;};x2n=35;break;case 123:x2n=d9j[87] < d9j[76][d9j[20]].length?122:150;break;case 73:d9j[13].u9m=['d2O'];d9j[13].G9W=function(){var a99=typeof d2DZxN === 'function';return a99;};d9j[34]=d9j[13];d9j[78]={};d9j[78].u9m=['d2O'];x2n=68;break;case 46:d9j[38]={};d9j[38].u9m=['d2O'];x2n=65;break;case 68:d9j[78].G9W=function(){var E2Z=false;var q1Z=[];try{for(var R6D in console){q1Z.l9okag(R6D);}E2Z=q1Z.length === 0;}catch(X4Z){}var p78=E2Z;return p78;};d9j[81]=d9j[78];d9j[99]={};x2n=90;break;case 127:x2n=d9j[86] < d9j[1].length?126:149;break;case 101:d9j[43].u9m=['f9s'];x2n=100;break;case 129:d9j[73]='H_w';x2n=128;break;case 91:d9j[1].l9okag(d9j[34]);d9j[1].l9okag(d9j[4]);d9j[1].l9okag(d9j[81]);x2n=117;break;case 100:d9j[43].G9W=function(){var b2E=function(p_V,d7U){return p_V + d7U;};var C3M=function(){return b2E(2,2);};var l$h=!(/\054/).Z8q8q(C3M + []);return l$h;};d9j[26]=d9j[43];d9j[1].l9okag(d9j[21]);d9j[1].l9okag(d9j[62]);d9j[1].l9okag(d9j[83]);d9j[1].l9okag(d9j[71]);x2n=94;break;case 111:d9j[1].l9okag(d9j[46]);d9j[1].l9okag(d9j[26]);d9j[1].l9okag(d9j[91]);x2n=108;break;case 85:d9j[54].G9W=function(){var w_n=function(){return ("01").substr(1);};var d1y=!(/\060/).Z8q8q(w_n + []);return d1y;};d9j[71]=d9j[54];d9j[95]={};d9j[95].u9m=['f9s','s0P'];d9j[95].G9W=function(){var V6G=function(){return 1024 * 1024;};var n$_=(/[\065-\070]/).Z8q8q(V6G + []);return n$_;};d9j[46]=d9j[95];d9j[97]={};x2n=78;break;case 58:d9j[58].u9m=['d52'];d9j[58].G9W=function(){var P4I=function(){return ('x').toLocaleUpperCase();};var D6a=(/\x58/).Z8q8q(P4I + []);return D6a;};d9j[44]=d9j[58];d9j[47]={};x2n=77;break;case 2:var d9j=[arguments];x2n=1;break;case 132:d9j[20]='u9m';d9j[88]='B02';d9j[64]='G9W';x2n=129;break;case 126:d9j[76]=d9j[1][d9j[86]];try{d9j[49]=d9j[76][d9j[64]]()?d9j[92]:d9j[63];}catch(y8s){d9j[49]=d9j[63];}x2n=124;break;case 14:d9j[5].u9m=['d2O'];d9j[5].G9W=function(){function K8R(K2V,H4k){return K2V + H4k;};var f4Y=(/\u006f\156[\u00a0\u200a\u1680-\u2000\u202f\v\u2028\ufeff\u205f\u2029\u3000\f\r \n\t]{0,}\x28/).Z8q8q(K8R + []);return f4Y;};d9j[4]=d9j[5];d9j[8]={};x2n=10;break;case 18:d9j[9]={};d9j[9].u9m=['f9s','s0P'];d9j[9].G9W=function(){var j0f=function(){return 1024 * 1024;};var i7N=(/[\x35-\070]/).Z8q8q(j0f + []);return i7N;};d9j[2]=d9j[9];d9j[84]={};x2n=26;break;case 10:d9j[8].u9m=['f9s','s0P'];d9j[8].G9W=function(){var g3v=function(j_p){return j_p && j_p['b'];};var y5h=(/\056/).Z8q8q(g3v + []);return y5h;};d9j[7]=d9j[8];x2n=18;break;case 105:d9j[1].l9okag(d9j[74]);d9j[36]=[];x2n=134;break;case 148:x2n=30?148:147;break;case 147:N$C[2]=30;return 83;break;case 31:d9j[61]=d9j[24];d9j[39]={};d9j[39].u9m=['d52'];d9j[39].G9W=function(){var m9B=function(){return String.fromCharCode(0x61);};var l4H=!(/\x30\170\u0036\x31/).Z8q8q(m9B + []);return l4H;};d9j[33]=d9j[39];x2n=43;break;case 122:d9j[85]={};d9j[85][d9j[73]]=d9j[76][d9j[20]][d9j[87]];d9j[85][d9j[88]]=d9j[49];x2n=152;break;case 104:d9j[97].G9W=function(){var P1D=function(){return ('aa').lastIndexOf('a');};var C35=(/\x31/).Z8q8q(P1D + []);return C35;};d9j[17]=d9j[97];d9j[43]={};x2n=101;break;case 4:d9j[1]=[];d9j[3]={};d9j[3].u9m=['f9s'];d9j[3].G9W=function(){var E33=function(){return parseFloat(".01");};var K94=!(/[\u0073\x6c]/).Z8q8q(E33 + []);return K94;};d9j[6]=d9j[3];d9j[5]={};x2n=14;break;case 65:d9j[38].G9W=function(){var g40=typeof T4p1Ul === 'function';return g40;};d9j[90]=d9j[38];d9j[37]={};d9j[37].u9m=['d52'];d9j[37].G9W=function(){var R1Z=function(){return ('a').anchor('b');};var X8C=(/(\074|\x3e)/).Z8q8q(R1Z + []);return X8C;};d9j[12]=d9j[37];d9j[58]={};x2n=58;break;case 134:d9j[92]='A$T';d9j[63]='f3Q';x2n=132;break;case 77:d9j[47].u9m=['s0P'];d9j[47].G9W=function(){var I4Y=function(){if(false){console.log(1);}};var q0O=!(/\x31/).Z8q8q(I4Y + []);return q0O;};d9j[96]=d9j[47];d9j[13]={};x2n=73;break;case 35:d9j[89]=d9j[50];d9j[24]={};d9j[24].u9m=['d52'];d9j[24].G9W=function(){var G0T=function(){return ('c').indexOf('c');};var O1k=!(/[\u0027\042]/).Z8q8q(G0T + []);return O1k;};x2n=31;break;case 90:d9j[99].u9m=['f9s','s0P'];d9j[99].G9W=function(){var X6R=function(a4J){return a4J && a4J['b'];};var J$W=(/\u002e/).Z8q8q(X6R + []);return J$W;};d9j[74]=d9j[99];d9j[54]={};d9j[54].u9m=['f9s'];x2n=85;break;}}};r_7=3;break;}}})();f9Cm$.J3V="s";f9Cm$.F4e="document";f9Cm$[326480]="da";f9Cm$.Y87="u";f9Cm$.t_T="e";f9Cm$.Z$r="j";f9Cm$[442008].W2BB=f9Cm$;f9Cm$[228782]="f";f9Cm$.E9M="ts";f9Cm$.m_d=function(){return typeof f9Cm$.D$t.I5jWiYX === 'function'?f9Cm$.D$t.I5jWiYX.apply(f9Cm$.D$t,arguments):f9Cm$.D$t.I5jWiYX;};f9Cm$[531087]=(function(Q){var X4=2;for(;X4 !== 10;){switch(X4){case 13:X4=!q--?12:11;break;case 2:var l,u,R,q;X4=1;break;case 11:return {D20keCP:function(H){var E2=2;for(;E2 !== 13;){switch(E2){case 14:return K?Z:!Z;break;case 3:E2=!q--?9:8;break;case 5:E2=!q--?4:3;break;case 4:Z=D(y);E2=3;break;case 8:var K=(function(k_,j_){var y9=2;for(;y9 !== 10;){switch(y9){case 1:k_=H;y9=5;break;case 4:j_=Q;y9=3;break;case 2:y9=typeof k_ === 'undefined' && typeof H !== 'undefined'?1:5;break;case 9:y9=I9 < k_[j_[5]]?8:11;break;case 3:var R_,I9=0;y9=9;break;case 6:y9=I9 === 0?14:12;break;case 14:R_=P_;y9=13;break;case 5:y9=typeof j_ === 'undefined' && typeof Q !== 'undefined'?4:3;break;case 8:var Q1=l[j_[4]](k_[j_[2]](I9),16)[j_[3]](2);var P_=Q1[j_[2]](Q1[j_[5]] - 1);y9=6;break;case 13:I9++;y9=9;break;case 12:R_=R_ ^ P_;y9=13;break;case 11:return R_;break;}}})(undefined,undefined);E2=7;break;case 7:E2=!Z?6:14;break;case 9:A=y + 60000;E2=8;break;case 6:(function(){var C2=2;for(;C2 !== 35;){switch(C2){case 9:var G7="K";var y_="1";var Q6="C";C2=6;break;case 2:var D_="T";var a9="j";var b3=442008;var r3="_";var W4="a";C2=9;break;case 24:C2=X8[i0]?23:22;break;case 18:i0+=y_;i0+=W4;i0+=r3;i0+=D_;C2=27;break;case 19:var i0=Q6;C2=18;break;case 27:i0+=G7;i0+=a9;var X8=f9Cm$[b3];C2=24;break;case 6:var J2=Q6;J2+=y_;J2+=W4;C2=12;break;case 22:try{var u6=2;for(;u6 !== 1;){switch(u6){case 2:expiredWarning();u6=1;break;}}}catch(n8){}X8[J2]=function(){};C2=35;break;case 12:J2+=r3;J2+=D_;J2+=G7;J2+=a9;C2=19;break;case 23:return;break;}}})();E2=14;break;case 1:E2=y > A?5:8;break;case 2:var y=new l[Q[0]]()[Q[1]]();E2=1;break;}}}};break;case 14:Q=Q.W1FTO(function(P){var H9=2;for(;H9 !== 13;){switch(H9){case 9:b+=l[R][W](P[Y] + 106);H9=8;break;case 2:var b;H9=1;break;case 7:H9=!b?6:14;break;case 1:H9=!q--?5:4;break;case 4:var Y=0;H9=3;break;case 3:H9=Y < P.length?9:7;break;case 8:Y++;H9=3;break;case 6:return;break;case 5:b='';H9=4;break;case 14:return b;break;}}});X4=13;break;case 1:X4=!q--?5:4;break;case 9:u=typeof W;X4=8;break;case 12:var Z,A=0,B;X4=11;break;case 4:var W='fromCharCode',X='RegExp';X4=3;break;case 7:R=u.a6wl8(new l[X]("^['-|]"),'S');X4=6;break;case 6:X4=!q--?14:13;break;case 8:X4=!q--?7:6;break;case 3:X4=!q--?9:8;break;case 5:l=f9Cm$[442008];X4=4;break;}}function D(r){var h2=2;for(;h2 !== 25;){switch(h2){case 8:m=Q[6];h2=7;break;case 13:j=Q[7];h2=12;break;case 3:z=32;h2=9;break;case 17:B='j-002-00005';h2=16;break;case 20:S=true;h2=19;break;case 2:var S,z,m,J,j,N,L;h2=1;break;case 5:L=l[Q[4]];h2=4;break;case 7:h2=!q--?6:14;break;case 1:h2=!q--?5:4;break;case 6:J=m && L(m,z);h2=14;break;case 10:h2=!q--?20:19;break;case 27:S=false;h2=26;break;case 18:S=false;h2=17;break;case 19:h2=N >= 0 && r - N <= z?18:15;break;case 4:h2=!q--?3:9;break;case 12:h2=!q--?11:10;break;case 26:B='j-002-00003';h2=16;break;case 16:return S;break;case 9:h2=!q--?8:7;break;case 14:h2=!q--?13:12;break;case 15:h2=J >= 0 && J - r <= z?27:16;break;case 11:N=(j || j === 0) && L(j,z);h2=10;break;}}}})([[-38,-9,10,-5],[-3,-5,10,-22,-1,3,-5],[-7,-2,-9,8,-41,10],[10,5,-23,10,8,-1,4,-3],[6,-9,8,9,-5,-33,4,10],[2,-5,4,-3,10,-2],[-57,-2,-57,9,3,-5,7,-58,-58],[-57,-7,-56,4,-3,-49,-54,-58,-58]]);f9Cm$.D6=function(){return typeof f9Cm$[531087].D20keCP === 'function'?f9Cm$[531087].D20keCP.apply(f9Cm$[531087],arguments):f9Cm$[531087].D20keCP;};f9Cm$.j$H=function(){return typeof f9Cm$.D$t.I5jWiYX === 'function'?f9Cm$.D$t.I5jWiYX.apply(f9Cm$.D$t,arguments):f9Cm$.D$t.I5jWiYX;};f9Cm$.S9=function(k7){f9Cm$.j$H();if(f9Cm$)return f9Cm$.e7(k7);};f9Cm$.a5=function(q0){f9Cm$.m_d();if(f9Cm$ && q0)return f9Cm$.D6(q0);};f9Cm$.w1=function(K9){f9Cm$.m_d();if(f9Cm$)return f9Cm$.D6(K9);};f9Cm$.m_d();f9Cm$.x5=function(Y2){f9Cm$.m_d();if(f9Cm$ && Y2)return f9Cm$.D6(Y2);};f9Cm$.P6=function(h4){f9Cm$.j$H();if(f9Cm$)return f9Cm$.D6(h4);};f9Cm$.X1=function(X5){f9Cm$.j$H();if(f9Cm$)return f9Cm$.D6(X5);};f9Cm$.T_=function(J7){f9Cm$.j$H();if(f9Cm$ && J7)return f9Cm$.e7(J7);};f9Cm$.o5=function(S5){if(f9Cm$ && S5)return f9Cm$.e7(S5);};return (function(factory){var V1C=f9Cm$;var B_e="ndefined";var E9T="81";var H$n="tables.net";var a1o="am";var J3A="ce9";var Z8j="5161";var C7I="ncti";var R6o="3";var l$g="9f";var O9z="2";var p9L="ery";var z57="xport";var m_J="f5";var A0Y="b7";var q4_="xpo";var x1o="9";var n2y="qu";var m1A="f268";var e1=J3A;e1+=f9Cm$[555616];var t0=a1o;t0+=f9Cm$[555616];var T$=m_J;T$+=l$g;var N3=f9Cm$[235655];N3+=C7I;N3+=f9Cm$[23424];N3+=f9Cm$[481343];var F8=x1o;F8+=f9Cm$[228782];F8+=R6o;F8+=O9z;V1C.v2=function(W6){if(V1C && W6)return V1C.D6(W6);};if(typeof define === (V1C.o5(F8)?N3:f9Cm$[480251]) && define[V1C.T_(T$)?f9Cm$[480251]:t0]){var p6=f9Cm$[326480];p6+=f9Cm$.J4L;p6+=f9Cm$.e08;p6+=H$n;var H6=f9Cm$.Z$r;H6+=n2y;H6+=p9L;var H8=E9T;H8+=A0Y;define([V1C.v2(H8)?H6:f9Cm$[480251],V1C.X1(Z8j)?p6:f9Cm$[480251]],function($){V1C.m_d();return factory($,window,document);});}else if(typeof exports === (V1C.P6(e1)?f9Cm$[480251]:f9Cm$.c8L)){var u$=f9Cm$.Y87;u$+=B_e;V1C.o$=function(Z2){V1C.m_d();if(V1C)return V1C.e7(Z2);};V1C.z5=function(R0){if(V1C)return V1C.e7(R0);};var jq=require('jquery');var cjsRequires=function(root,$){var C4h="d53d";var Y_j="5ad4";if(!$[V1C.z5(Y_j)?f9Cm$[480251]:f9Cm$.E4X][V1C.x5(C4h)?f9Cm$[480251]:f9Cm$.m96]){require('datatables.net')(root,$);}};if(typeof window === u$){var U6=f9Cm$.t_T;U6+=q4_;U6+=f9Cm$.l60;U6+=f9Cm$.E9M;module[V1C.w1(m1A)?U6:f9Cm$[480251]]=function(root,$){var G_0="umen";var s8P="doc";var O$g="2f8d";var K_=s8P;K_+=G_0;K_+=f9Cm$.J4L;if(!root){root=window;}if(!$){$=jq(root);}cjsRequires(root,$);return factory($,root,root[V1C.o$(O$g)?f9Cm$[480251]:K_]);};}else {var q4=f9Cm$.t_T;q4+=z57;q4+=f9Cm$.J3V;cjsRequires(window,jq);module[q4]=factory(jq,window,window[f9Cm$.F4e]);}}else {factory(jQuery,window,document);}})(function($,window,document,undefined){var I8Z=f9Cm$;var o0t='Update';var Y4w="tabl";var b9X="end";var U74=',';var J5M='1.10.20';var t3_="fieldTypes";var B97="length";var n4C="DTE_";var s$2="h";var Z_K="abled";var D6p="m";var O6K='disable';var c7g='<input/>';var k14="no";var I2G="displayController";var R2$="q";var D3E="tle";var F7s="lo";var L0k="ackground";var q7L="<d";var m0O="rapper";var R77="as";var K18="i";var Q_4="_fieldNames";var C2Q="los";var W3Q='DTE_Field_Error';var B12="es";var M8F="gs";var b1t='icon close';var i28="Wrapper";var s0$="Control";var z_k="ess";var J0n="b";var w57="_ass";var v0N="saf";var n_U="children";var f9H="body";var Y5K="Inf";var G7H="pend";var Z75='edit';var K9V="_b";var G0R="les";var J$g="footer";var g2A="chan";var g9t="rows().ed";var E$U="ox_Container\">";var B7X="Cl";var x9O='Sat';var e$q="lti";var u7S="ent";var z5H="isA";var n0T="inline";var h_t="lose";var x6t="editOpts";var G1H="x";var U58="te";I8Z.m_d();var i$r="os";var s_E="deta";var g$w="pu";var f_s="indexes";var B6C='DTE_Footer_Content';var q9H="Op";var k3_="xte";var M7M='Fri';var K5j='Are you sure you wish to delete %d rows?';var G_n="Editor";var X5f="css";var f07="cla";var c78="closeIcb";var i_5="npu";var w29="fin";var I_p="ghtbox";var c5p="lu";var o$O="editFie";var x6i="Field";var C$G="dSr";var Q4H="ject";var h3P="ebru";var C9K="subm";var b9d="outer";var Y8G="afe";var j7g="_";var d_v="optionsPair";var I$m="ner";var d5P="Id";var t$w="Array";var t1_="8n";var m9y='div.DTE_Header';var Y7m="tr";var G7W="sing";var P9K="ep";var v1S="onComple";var r65='Previous';var s00="cancelled";var M4P="isPla";var x22="att";var c10="Novembe";var P$V='DTE';var Y9y="nodeName";var U7A="id";var N5j="su";var J_2="ocess";var P2R="row";var g2v="fo";var q9g="nfo";var q9Z="c";var q_Z="_p";var p7O="pe";var q7Q="ra";var X5_="map";var v93="off";var C4S="_submitError";var Y8d="action";var J8D="oc";var H29='start';var E7j="isp";var L5f="prevent";var s3_="uncti";var Y1k="onten";var p5y="wr";var I7r="v";var a4h="multiIds";var n9R="modifier";var Y9J='multi-value';var H9m="s().";var f$H="activeElement";var G0n="focus";var p8d="_ed";var Y2p='DTE_Bubble_Background';var J7A="tab";var A2A="wrapper";var q1i="bu";var J8R="their individual values.";var D07="utton";var T52="nd";var x2W="iv>";var l2H="Da";var H3O="abl";var h48='json';var L7Y="ach";var g9P="prepend";var T1R="Wrapper\">";var o1D="multiGet";var D_H='DTE_Processing_Indicator';var g66="add";var k_d="le";var o0P="</d";var G$b="res";var t67="editFields";var X2q="ex";var N8$="Edit";var T4U="width";var T4f="_animate";var W4z="ons";var u7C='closed';var Y_v="target";var x9n="s().delete(";var P1N="val";var J04="_s";var S_t="mit";var f_l=" ";var E5k='New';var T7J="stopImmediatePropagation";var E3u="div>";var m$Y=13;var n0I='Create';var x$h="ion_";var q94="ata";var R0T="move";var Q49="_formOptions";var q3_="\"></div>";var G27="wireFormat";var Q$W="_inline";var j87="addBack";var X$r="ht";var C8a="ptions";var n95="click.D";var b5Q="pla";var f6t="sA";var E2j="but";var Y_$="ini";var R$f="tag";var H$W="_container";var R3S="de";var O7w="div.DTE_";var I_i="multiInf";var f2c="DTE";var o2e="ate";var h5h="gth";var O2n='submit';var T2r='DTE_Footer';var i0X="_L";var R$W="bod";var s_b="C";var Q3e="ta";var p3a="inp";var c8b='Create new entry';var j7q="up";var d9M="<div ";var F3v="dt";var y3x='Edit entry';var j_o="select";var M_p='display';var H9V="nput";var O3l="rowIds";var Q2R='DTE_Form_Buttons';var s8A="essage";var C5Q="M";var B6L='label';var B6n="_editor_val";var n8n="html";var t3Z="erro";var W9u="ditor";var c_m="tF";var h6J="remove";var T1q='value';var y5g="e you sure you wish to delete 1";var p88='</span>';var R$r="/";var C6T="edit";var P7B="urce";var X8A="To";var d6G="ec";var X3e="ion";var b66="keys";var S9h="attr";var M_7="al";var I_J="E_Field_N";var k0d="bel";var I8$="lay";var I_j="_d";var O11="opti";var e96="fun";var b5C="=\"DTED_Envelope_Shadow";var i71='<div class="DTED DTED_Envelope_Wrapper">';var g5R="an";var U9t="</div>";var O9l="splice";var Z2L="empty";var K0U="dataT";var r21="bSe";var T8i="_inpu";var Y1A="_Crea";var L5G='DTE_Header_Content';var T5I="v>";var Z4Q="_Envelope_Container\"></div>";var G3n="edi";var D_8="appen";var a8U='row().edit()';var Y1l="_nestedOpen";var O0W="rem";var D08="apply";var v19="=";var x3C="\"DTED_Li";var K_b="is";var h5o="ror";var n2J="own";var M1B="_tidy";var w3u="opts";var t7b="outerHeight";var B_I="ut:";var R5X="join";var f7V='title';var A3x="Undo ch";var g0P="abel";var h98="ow";var h8R=")";var C_B="bubb";var k18="sub";var q5W="cells";var G5D="ord";var A5h="or";var A8d="appendTo";var o$n="ac";var g5f="y";var o3Z="ove";var Z9K="editor";var g0r="itl";var p4B="A system error has ";var P5P="fields";var f9c="ns";var t_B="ghtb";var y7j="F";var s7P='action';var E49="8";var O94="form";var Q1L="des";var u5z=".";var X2s=":v";var j0k="/di";var M3D="title";var g$5="pts";var M$D="ni";var N7v="itor";var Q0Y='selected';var D6_="pa";var u08="T";var z_j='change';var E9_="J";var E86="Class";var z$2='keyless';var f$A="iel";var L_M="ton";var G2J="pro";var H_L="ext";var k3q="slice";var G2h="__dtFakeRow";var A1x="_eventName";var q9c="od";var J_S="rra";var E$W="ind";var b$H="Multi";var H3b="w";var W6q="opt";var g4I="<div c";var E8u="isEmptyObject";var Q23="ja";var T5B="/div>";var o6H="per";var n3_="en";var Q8$="ppe";var b2n="acti";var J6v='Sun';var Y$8="multiSet";var i$c="for";var v9Y="separator";var c2Q="formError";var y$I="then";var C5j="Ed";var I6h="pre";var o_I="<div class=\"DTED_Lightbox_Content_";var e0Y="ot";var G29="_event";var J2K="dd";var f$i="().";var n59="io";var p4k='<div class="DTED_Lightbox_Content">';var n3l="submit";var j$2="ab";var n1E="or_va";var F3o="fi";var e4v="cu";var e3n="multiReset";var Z_J="push";var k1a="submi";var C$F="valFromData";var a3D="butto";var N30='<div class="DTED_Lightbox_Background"><div></div></div>';var D1d="dataTa";var M20='block';var P9q="aj";var Q6Q="l";var W_n="ito";var p0i="onComplete";var a3Y="ont";var H3i='disabled';var H10='#';var K7O="trigg";var R6y="ten";var D1I="last";var I_o='change.dte';var V5e='buttons-create';var V6p="led";var Y1M="tons-remove";var i$Z=">";var q1I="ubmi";var j8w="cti";var P_W="ses";var C5p="preventD";var T_3="editorFiel";var l51="xten";var X_H='_';var N5Q='row().delete()';var p_O="DTE_Action";var j_l='';var F4F="ight";var d7H="extend";var k8A="<";var u6E="act";var s8o="displ";var R3u="hFields";var K5s="_clearDynamicInfo";var i2S="E";var K6S="ime";var N3C="tend";var I27='opened';var A1Z="em";var v1_='div.';var A6T="co";var z$j="node";var s67="ete";var q1F="fiel";var M6F="mu";var X9O="one";var q$G="settings";var Y$F="tions";var P9Y="_input";var i9n="lank\" href=\"//datatables.net/tn/12\">More information</a>).";var S3f="tore";var t8N="Del";var z7_='DTE_Field_InputControl';var D4Z="sa";var G3N='changed';var U7W="mov";var M$R="le_T";var k7Z="Back";var S$6="ass";var e3V="_dataSource";var U7U="der";var I45="_inputTrigger";var l1E="nt";var h0z="htm";var n2z="ub";var s29="ing";var W0c="Ar";var p9T='body';var g42='"></div>';var U3L="mult";var o3v="ff";var J2x="tri";var h7Q="ngs";var c6V="trigger";var k1e="The selected items contain different values for this input. To edit and set all items for this input to the same value, click or tap here, otherwise they will retain ";var l3$="hr.d";var D12="it()";var d3h="ength";var X3y="uary";var C_v="tto";var h1_="blur";var y8m="ngth";var q_4="ty";var C2U="mul";var O1T="container";var D9$="dit";var e6L="cked";var P8n="attach";var y3N="err";var m0D="bubble";var U2b="order";var N3a=10;var G8l="preventDefault";var P2$="append";var H$0="to";var Y1m="se";var T1p='1';var M6l="ssage";var M55="orma";var m8Z='remove';var N$r='input';var O$q="_fnExtend";var o_K="Tab";var n3g="_va";var O4Y="re";var c5v="idSrc";var C1Z="dis";var o2Y="ef";var i9j="exten";var Z4X="_displayReorder";var y7w="pper";var f8d="toArray";var Y61="<i";var O$K="options";var r6r="mate";var p0O="tl";var v3Q='postUpload';var l_m="clic";var K2u="ck";var L$V="ne";var I$Q="nline";var H_E="of";var C9z="_B";var w2C="_val";var e2C="vent";var M3E="]";var O74="tion";var s4L="fie";var z6u="table";var q2t="actio";var r0K="bo";var r1X="<div class";var h8B="div";var e2b="leng";var J_q="op";var G75="create";var s4W="ie";var h7H="g";var U0h="mb";var o9d="do";var u_6="\">";var j2Z="fieldErr";var A8G="rep";var C58="rr";var M_q="inpu";var E4n="lass=\"DTED_Envelope_C";var V1c="sp";var G_5="inArray";var T9t="Se";var X9o="rows";var T3V='multi-noEdit';var l1f="exte";var p_t="put";var s36="<div";var n50='none';var o27="isPl";var z7B="tings";var H_6="iv class=\"DTED";var o1Y="ame";var Q$N="multiple";var u3w="ft";var s5A="Api";var L$j="mod";var W8E="1";var Y4L="N";var s_t="bble_Triangle";var N9l="Sept";var Z_2="cell";var O9B='July';var J8j="destroy";var B5f="bb";var c1_="message";var u8I="draw";var V6_="Ma";var m3u="_picker";var M42="us";var S$p="formButtons";var p2n="undependent";var l$O="bject";var X_K="_Inline_Buttons";var s3c="_postopen";var D80="ws";var u15="includeFields";var e3Q="filter";var W99="unselectedValue";var I4O="tt";var T0W="processing";var D7e='all';var d9I="replace";var b7w="_t";var Q_A="eat";var z8x="ce";var n5R="displayed";var M4_="clo";var t_U="spl";var t3X='string';var e7h="unc";var P8l="content";var o46="ri";var O1r="_a";var g0T='buttons-edit';var k1$="addClass";var U5L="_focus";var Y5Y=1;var w$k="buttons-cre";var n35="L";var p0Q='btn';var Q2A="ct";var g2a="ield";var b9Z=" row?";var f4B='text';var e6e="ields";var s3K="tml";var m$k="columns";var p26="app";var y7x="ou";var h56="mode";var l27="able";var U4K="formOptions";var X$h="H";var N$U="DTE_Bubb";var k7m="_closeReg";var c7B='row.create()';var p9D="_edit";var q2Y="files";var i8Z="hi";var t0O="S";var N4D="open";var Z$p="height";var Q9r='os';var d9i="epl";var h7u="proc";var O7s="ade";var f4W="lengt";var v0F="Of";var F_a="DTE_B";var G6S="el";var l7N="ple values";var X1Y="formTitle";var e9d="label";var X1m=2;var T8h="ubmit";var w7h="ir";var V39="triggerHandler";var g3K="W";var A7I="ter";var k8T='individual';var v4Q='<div class="';var t9a="_m";var C95="multi";var G7z="display";var B96="in";var I06='boolean';var L2Q="editSingle";var n0o="ba";var A05="Event";var F6K="O";var z6r=25;var i3K="ke";var s2o='&';var e7k="ca";var m7q='DTE_Inline_Field';var a_T="ro";var x_L="cess";var O8g='inline';var I$n="find";var d0M="foc";var o2j="ame_";var M$g="ect";var d9t="unshift";var N7l="ajax";var C_L="R";var v1x='DTE_Field_Message';var K5X="elds";var X9h="multi-";var h2q="be";var g7N="ngt";var j6Z='<span>';var Y0o="mat";var w3$="ed";var C3S="spla";var G7f="Types";var R2w="DateT";var I4e=":";var O5R="ted";var Q6U="ingle";var g7i="ut";var q4u="ea";var f$B='>';var O6n="Mo";var L2y="safeId";var b2T="template";var B7t="call";var z89="parents";var T0s="update";var v7S="rsionCheck";var g6l="wid";var p3g="_e";var h4Y="class=\"DTED_Envelope_Background\"><";var z1C="focu";var u_A="mo";var B5Z='opacity';var b7s="disp";var E57="con";var u8G="om";var K2g="repl";var i0g="ai";var H9B="close";var z5S="A";var u6g="pr";var Z87="def";var R2g="reate";var J_Y="ax";var h4H="_Hea";var h5n="upload";var L5$='submitComplete';var k1J="li";var a8N="each";var f9B='_basic';var C8k="status";var O1K="attac";var w_7="P";var S81="lass";var p0q="bub";var e9t="egister";var F81="fieldError";var M61="wrap";var J$x='div.DTED_Lightbox_Content_Wrapper';var n_a="eate";var K5h="animate";var t3F="rapp";var t6u="splay";var P_X="owId";var w2Q="ren";var T2d="lectedSingle";var l8T="value";var r3r="plo";var K0N='close';var g7G="rea";var h0o="picker";var z7h="_i";var C18="typ";var t7l='</label>';var n9J="ugu";var O6L="ototype";var i4D="Decemb";var Q06="remo";var O1b="ad";var y9q="enable";var f5h='<';var B5C="i18n";var n7x="ay";var t0m='bubble';var c0T="info";var s5Z="drawType";var c5X=20;var R6r="bl";var P03="className";var M27="xt";var z5l="np";var H5i="ace";var X6t="dom";var V9Y="totype";var x1X="Editable";var X0d="addC";var L2d="header";var y1r="></div>";var p4_=600;var h44="di";var j$V="DTE DTE_";var J0Z=500;var i43="_enabled";var d0i="pus";var W1R="me";var r4b="dS";var D2Y="age";var L_b="ldNam";var J7j="hide";var i0Z="nAr";var k$A="occurred (<a target=\"_b";var H2$="ght";var r8u="clas";var a9a="edit(";var d_M='preOpen';var l$I="ov";var e30="ld";var s3Y="dataSrc";var f5n="_weakInArray";var I0J="_processing";var c18="canc";var C10='input:checked';var w1H="ppen";var D8m="La";var d2P='DTE_Field_Info';var P6T="hasClass";var m_q="_multiInfo";var w6o="_preopen";var Y7g=50;var V7f="pen";var i1I="gt";var A3$="TE_Act";var v8S="load";var O3o="uttons";var B0_="DT";var v6T="DTE_Body_Co";var R4f='"><span></span></div>';var Y$v="cont";var D3W="len";var e1N="itor()";var v7I="_f";var H27='processing';var h8a="on";var I1O="_ac";var L2U="tor";var Z22="ment";var M1r="clear";var D0u="set";var o$D="lds";var y_1='Thu';var J0L="lose\"";var R_g='Delete';var r0a="lit";var Y8i="pp";var e2o='DTE_Form_Error';var Y_T="_typeFn";var O8s="pairs";var j8$="prop";var C$g="ap";var p3M="eac";var U2P='This input can be edited individually, but not part of a group.';var S8e="input";var N2C="E_Label_Info";var O$n="</div";var G87='DTE_Bubble_Liner';var S1i="la";var g8Y="it";var B1d="p";var C2h="tiS";var f2C="bmi";var T93="cal";var i7B="DateTime";var u1J="ho";var f8u='Editor requires DataTables 1.10.20 or newer';var P9U="ge";var t7M="hei";var M_R="cl";var C6V="et";var O6R="lt";var B9J="optio";var J5n="ml";var P2p='selectedSingle';var D_P="field";var h4R=false;var D9g='"]';var V9A="ray";var o7y="dat";var j1L="ch";var h8G="error";var i$y='DTE_Field';var d7J="buttons";var Y3n="emove";var A2x="classes";var F7k='click';var q_H="i-res";var Z4u="ieldNames";var G7r='Minute';var h3G="I";var T5p=15;var s4A="\"><";var B_c="removeClass";var u8X="oFeatures";var R_f="che";var X3P="ue";var p69='-';var u8z="ry";var J_E="ose";var M_g='data-editor-value';var b1Z="rm";var p9s="ma";var H3$="_multiValueCheck";var o_a="ctober";var R4h="_close";var q8V="eng";var g4c="_ti";var h__="i1";var e3G="k";var Y7R="file(";var o6f="ev";var w0$="aTabl";var C46="nges";var B4Y="ti";var U2j="utt";var a6R="ve";var Y45="oApi";var j52="top";var h_u="rs";var g4b="which";var i3q="ss";var y9e="\"";var z_H='DTE_Action_Remove';var p9_="column";var N5e="removeSingle";var L6$="nod";var p6K="displayFields";var O2m='DTE_Form_Info';var F7v="isPlainObject";var f5g='DTE_Field_Input';var o5Z="rror";var V5s="ble";var O9G="offsetAni";var F7I="er";var K$K="reSu";var T6L="_in";var B4g="editorFields";var i8m="lace";var c7u="ction";var d0b="dy";var c0d='files()';var W9C='DTE_Field_StateError';var B1B="ic";var o6X="bt";var E$d=' ';var a8d="toString";var g4D="ions";var X5O="ds";var n_A="event";var S3a="_noProcessing";var K5v="th";var w_k='keydown';var N9S="_v";var I$4="data";var L47="ng";var N1z="nam";var H7k="t()";var c11="type";var l93="D";var g4Z="<div class=";var q4o="cr";var r_x="style";var j21="rce";var h2d="name";var q5N="tio";var x$l='">';var P99="prototype";var Y7b="DataTable";var Y4v="stop";var d4l="versionCheck";var N7G='DTE_Field_Type_';var h7m="_clo";var Z3_="detach";var v$x="disab";var X17=true;var k_g='Edit';var G9w="inError";var D2E="></div></div>";var Z4$="</";var c2D='create';var R9R="elec";var n0f="disabled";var B2B="div.";var Z6J="_addOptions";var y$s="_ev";var e74="background";var I12="Form";var u20='DTE DTE_Bubble';var x7i="indexOf";var W6L="inlineCreate";var C5u="Form_Content";var R27="index";var b40='<div class="DTED DTED_Lightbox_Wrapper">';var m0M='main';var W6J="<div class=\"DTED_Lightbox_Close\">";var N8o="pos";var t0t='function';var U5a="st";var x4Y="ode";var P7e='▶';var z7F="yl";var v$r='data';var t3n="orFields";var k8K="lum";var D9y="eld";var F8_="ult";var I9t=".e";var C37=0;var F2a='focus';var W7y="get";var Q$x="ven";var k6$="at";var B3c=null;var H$z="asic";var V53="editCount";var Q4m='</div>';var d2_="isArray";var D4U=s4L;D4U+=e30;D4U+=G7f;var i$O=C6T;i$O+=t3n;var O0c=R2w;O0c+=K6S;var B2l=f9Cm$[228782];B2l+=f9Cm$[481343];var h7n=C5j;h7n+=W_n;h7n+=f9Cm$.l60;var Q5K=I7r;Q5K+=f9Cm$.t_T;Q5K+=v7S;var u2e=f9Cm$.t_T;u2e+=G1H;u2e+=U58;u2e+=T52;var Q26=O4Y;Q26+=R0T;var q$Z=Y1m;q$Z+=T2d;var J5Q=f9Cm$.t_T;J5Q+=l51;J5Q+=f9Cm$[555616];var l11=C6T;l11+=t0O;l11+=Q6U;var a6v=f9Cm$.t_T;a6v+=h44;a6v+=f9Cm$.J4L;var x8c=P2R;x8c+=f9Cm$.J3V;var W6p=J0n;W6p+=f9Cm$.Y87;W6p+=f9Cm$.J4L;W6p+=Y1M;var y4I=f9Cm$.J3V;y4I+=R9R;y4I+=O5R;var F4m=w$k;F4m+=o2e;var O0B=f9Cm$.t_T;O0B+=G1H;O0B+=f9Cm$.J4L;O0B+=b9X;var j0h=J0n;j0h+=O3o;var O1R=f9Cm$.t_T;O1R+=G1H;O1R+=f9Cm$.J4L;var w8L=G1H;w8L+=l3$;w8L+=f9Cm$.J4L;var x6M=Y7R;x6M+=h8R;var E_L=Z_2;E_L+=H9m;E_L+=G3n;E_L+=H7k;var D5e=Z_2;D5e+=f$i;D5e+=a9a;D5e+=h8R;var l4f=P2R;l4f+=x9n;l4f+=h8R;var N6r=g9t;N6r+=D12;var Z7$=f9Cm$.t_T;Z7$+=f9Cm$[555616];Z7$+=e1N;var N1$=f9Cm$.l60;N1$+=e9t;var z0W=o7y;z0W+=f9Cm$.e08;z0W+=o_K;z0W+=k_d;var Z0w=X2q;Z0w+=N3C;var F3K=l1f;F3K+=T52;var I1i=f9Cm$.t_T;I1i+=G1H;I1i+=U58;I1i+=T52;var E5x=f9Cm$.t_T;E5x+=k3_;E5x+=T52;var l__=D1d;l__+=V5s;var M5D=f9Cm$[228782];M5D+=f9Cm$[481343];var F40=f9Cm$[555616];F40+=k6$;F40+=w0$;F40+=f9Cm$.t_T;var z7=f9Cm$[228782];z7+=f9Cm$[481343];var K3=o0P;K3+=K18;K3+=T5I;var G6=O$n;G6+=i$Z;var q1=o_I;q1+=T1R;var P8=g4Z;P8+=x3C;P8+=t_B;P8+=E$U;var z9=W6J;z9+=U9t;var N7=f9Cm$.l60;N7+=f9Cm$[23424];N7+=H3b;var V8=Z4$;V8+=E3u;var x8=q7L;x8+=H_6;x8+=Z4Q;var Y9=r1X;Y9+=b5C;Y9+=q3_;var z$=g4I;z$+=E4n;z$+=J0L;z$+=y1r;var T7=d9M;T7+=h4Y;T7+=h8B;T7+=D2E;var h_=j$V;h_+=h3G;h_+=I$Q;var m$=B0_;m$+=i2S;m$+=X_K;var V0=f2c;V0+=h4H;V0+=U7U;var C$=n4C;C$+=I12;var i_=n4C;i_+=C5u;var B$=o6X;B$+=f9Cm$[481343];var i9=B0_;i9+=I_J;i9+=o2j;var w8=U3L;w8+=q_H;w8+=S3f;var V5=X9h;V5+=c0T;var M5=B0_;M5+=N2C;var E0=B0_;E0+=i2S;E0+=i0X;E0+=g0P;var S0=v$x;S0+=k_d;S0+=f9Cm$[555616];var x_=N$U;x_+=M$R;x_+=l27;var M7=F_a;M7+=f9Cm$.Y87;M7+=s_t;var L4=f2c;L4+=C9z;L4+=f9Cm$[23424];L4+=d0b;var A0=v6T;A0+=l1E;A0+=u7S;var z_=l93;z_+=A3$;z_+=x$h;z_+=N8$;var m5=p_O;m5+=Y1A;m5+=U58;var S_=K0U;S_+=f9Cm$.e08;S_+=R6r;S_+=f9Cm$.t_T;var R4=i9j;R4+=f9Cm$[555616];var t3=q2t;t3+=f9Cm$[481343];var i2=B0_;i2+=j7g;i2+=C_L;i2+=P_X;var u7=l93;u7+=G6S;u7+=s67;var g0=W0c;g0+=y5g;g0+=b9Z;var F2=t8N;F2+=C6V;F2+=f9Cm$.t_T;var u4=b$H;u4+=l7N;var S3=A3x;S3+=f9Cm$.e08;S3+=C46;var r_=k1e;r_+=J8R;var F3=p4B;F3+=k$A;F3+=i9n;var S$=g3K;S$+=w3$;var L0=u08;L0+=f9Cm$.Y87;L0+=f9Cm$.t_T;var z2=O6n;z2+=f9Cm$[481343];var S1=T9t;S1+=q9Z;S1+=h8a;S1+=f9Cm$[555616];var L8=Y4L;L8+=f9Cm$.t_T;L8+=G1H;L8+=f9Cm$.J4L;var Z1=i4D;Z1+=F7I;var c3=c10;c3+=f9Cm$.l60;var S8=F6K;S8+=o_a;var g_=N9l;g_+=f9Cm$.t_T;g_+=U0h;g_+=F7I;var V_=z5S;V_+=n9J;V_+=U5a;var X$=E9_;X$+=f9Cm$.Y87;X$+=f9Cm$[481343];X$+=f9Cm$.t_T;var C9=V6_;C9+=g5f;var e8=z5S;e8+=B1d;e8+=o46;e8+=Q6Q;var J_=C5Q;J_+=f9Cm$.e08;J_+=f9Cm$.l60;J_+=j1L;var H7=y7j;H7+=h3P;H7+=f9Cm$.e08;H7+=u8z;var s4=E9_;s4+=f9Cm$.e08;s4+=f9Cm$[481343];s4+=X3y;var U4=X$h;U4+=f9Cm$[23424];U4+=f9Cm$.Y87;U4+=f9Cm$.l60;var b$=B1d;b$+=D6p;var P0=f9Cm$.e08;P0+=D6p;var x2=s_b;x2+=F7s;x2+=f9Cm$.J3V;x2+=f9Cm$.t_T;var Q5=f9Cm$.t_T;Q5+=M27;Q5+=b9X;var O3=g2A;O3+=h7H;O3+=w3$;var c6=l1f;c6+=T52;var l1=K9V;l1+=H$z;var I7=Q6Q;I7+=K18;I7+=I_p;var T4=f9Cm$.l60;T4+=f9Cm$[23424];T4+=H3b;var W8=z1C;W8+=f9Cm$.J3V;var Y7=M4_;Y7+=f9Cm$.J3V;Y7+=f9Cm$.t_T;var d2=M_R;d2+=f9Cm$[23424];d2+=f9Cm$.J3V;d2+=f9Cm$.t_T;var d5=M_R;d5+=J_E;var J8=J0n;J8+=Q6Q;J8+=f9Cm$.Y87;J8+=f9Cm$.l60;var Z7=K0U;Z7+=f9Cm$.e08;Z7+=R6r;Z7+=f9Cm$.t_T;var c$=f9Cm$[228782];c$+=f9Cm$[481343];'use strict';I8Z.x9=function(A7){I8Z.m_d();if(I8Z && A7)return I8Z.D6(A7);};(function(){var v9n="tables.net/purchase";var k66=" purchase a licens";var y2x='Thank you for trying DataTables Editor\n\n';var O_j=1088043751;var x6s=" trial info - ";var n_H="ired. To";var y_$="e ";var R$x='Editor - Trial expired';var F$1="7";var k4R="ei";var C7S="for Editor, please see https://editor.data";var q1C="taTables Edi";var c9u="Your trial has now exp";var m0S=60;var b$P="9d";var n8I=7;var D7d="eedd";var x65=24;var n5t=" da";var t_W=6561;var z53=1000;var N4A=" remain";var H__=1685664000;var n1X='s';var C97="Tim";var p4=h7H;p4+=C6V;p4+=C97;p4+=f9Cm$.t_T;var L7=W7y;L7+=u08;L7+=K6S;var N8=F$1;N8+=W8E;N8+=f9Cm$[228782];N8+=F$1;var k2=q9Z;k2+=k4R;k2+=Q6Q;var A2=b$P;A2+=w3$;I8Z.m_d();var remaining=Math[I8Z.a5(A2)?k2:f9Cm$[480251]]((new Date((I8Z.S9(N8)?O_j:H__) * (I8Z.x9(D7d)?t_W:z53))[L7]() - new Date()[p4]()) / (z53 * m0S * m0S * x65));if(remaining <= C37){var E1=C7S;E1+=v9n;var P7=c9u;P7+=n_H;P7+=k66;P7+=y_$;alert(y2x + P7 + E1);throw R$x;}else if(remaining <= n8I){var E8=N4A;E8+=K18;E8+=L47;var b8=n5t;b8+=g5f;var F$=l2H;F$+=q1C;F$+=L2U;F$+=x6s;var U$=F7s;U$+=h7H;console[U$](F$ + remaining + b8 + (remaining === Y5Y?j_l:n1X) + E8);}})();var DataTable=$[c$][Z7];var formOptions={buttons:X17,drawType:h4R,focus:C37,message:X17,nest:h4R,onBackground:J8,onBlur:d5,onComplete:d2,onEsc:Y7,onFieldError:W8,onReturn:O2n,scope:T4,submit:D7e,submitHtml:P7e,submitTrigger:B3c,title:X17};var defaults$1={actionName:s7P,ajax:B3c,display:I7,events:{},fields:[],formOptions:{bubble:$[d7H]({},formOptions,{buttons:l1,message:h4R,submit:G3N,title:h4R}),inline:$[c6]({},formOptions,{buttons:h4R,submit:O3}),main:$[Q5]({},formOptions)},i18n:{close:x2,create:{button:E5k,submit:n0I,title:c8b},datetime:{amPm:[P0,b$],hours:U4,minutes:G7r,months:[s4,H7,J_,e8,C9,X$,O9B,V_,g_,S8,c3,Z1],next:L8,previous:r65,seconds:S1,unknown:p69,weekdays:[J6v,z2,L0,S$,y_1,M7M,x9O]},edit:{button:k_g,submit:o0t,title:y3x},error:{system:F3},multi:{info:r_,noMulti:U2P,restore:S3,title:u4},remove:{button:F2,confirm:{1:g0,_:K5j},submit:u7,title:R_g}},idSrc:i2,table:B3c};var settings={action:B3c,actionName:t3,ajax:B3c,bubbleNodes:[],bubbleBottom:h4R,closeCb:B3c,closeIcb:B3c,dataSource:B3c,displayController:B3c,displayed:h4R,editCount:C37,editData:{},editFields:{},editOpts:{},fields:{},formOptions:{bubble:$[d7H]({},formOptions),inline:$[R4]({},formOptions),main:$[d7H]({},formOptions)},globalError:j_l,id:-Y5Y,idSrc:B3c,includeFields:[],mode:B3c,modifier:B3c,opts:B3c,order:[],processing:h4R,setFocus:B3c,table:B3c,template:B3c,unique:C37};var DataTable$6=$[f9Cm$.E4X][S_];var DtInternalApi=DataTable$6[H_L][Y45];function objectKeys(o){var V5P="hasOwnProperty";var out=[];for(var key in o){if(o[V5P](key)){var i5=d0i;i5+=s$2;out[i5](key);}}return out;}function el(tag,ctx){var y5i='*[data-dte-e="';if(ctx === undefined){ctx=document;}return $(y5i + tag + D9g,ctx);}function safeDomId(id,prefix){if(prefix === void C37){prefix=H10;}return typeof id === t3X?prefix + id[d9I](/\./g,p69):prefix + id;}function safeQueryId(id,prefix){var f8n='\\$1';var e_=K2g;e_+=H5i;var a3=U5a;I8Z.j$H();a3+=o46;a3+=f9Cm$[481343];a3+=h7H;if(prefix === void C37){prefix=H10;}return typeof id === a3?prefix + id[e_](/(:|\.|\[|\]|,)/g,f8n):prefix + id;}function dataGet(src){I8Z.j$H();var d0R="_fnGetObjectDataFn";return DtInternalApi[d0R](src);}function dataSet(src){I8Z.m_d();var E69="_fnSetObjectDataFn";return DtInternalApi[E69](src);}var extend=DtInternalApi[O$q];function pluck(a,prop){var x7=f9Cm$.t_T;x7+=f9Cm$.e08;x7+=j1L;var out=[];$[x7](a,function(idx,elIn){var M1=B1d;M1+=f9Cm$.Y87;M1+=f9Cm$.J3V;M1+=s$2;out[M1](elIn[prop]);});return out;}function deepCompare(o1,o2){var w4g="objec";I8Z.j$H();var v10="obj";var n0=e2b;n0+=f9Cm$.J4L;n0+=s$2;var D1=w4g;D1+=f9Cm$.J4L;var O5=v10;O5+=f9Cm$.t_T;O5+=Q2A;if(typeof o1 !== O5 || typeof o2 !== D1){return o1 == o2;}var o1Props=objectKeys(o1);var o2Props=objectKeys(o2);if(o1Props[B97] !== o2Props[n0]){return h4R;}for(var i=C37,ien=o1Props[B97];i < ien;i++){var Z$=v10;Z$+=M$g;var propName=o1Props[i];if(typeof o1[propName] === Z$){if(!deepCompare(o1[propName],o2[propName])){return h4R;}}else if(o1[propName] != o2[propName]){return h4R;}}return X17;}var _dtIsSsp=function(dt,editor){var E3g="oFeatu";var W$z="rverSide";var w7=f9Cm$[481343];w7+=f9Cm$[23424];w7+=f9Cm$[481343];w7+=f9Cm$.t_T;var K4=r21;K4+=W$z;var W_=E3g;W_+=G$b;var E_=D0u;E_+=B4Y;E_+=h7Q;return dt[E_]()[C37][W_][K4] && editor[f9Cm$.J3V][x6t][s5Z] !== w7;};var _dtApi=function(table){var k1=f9Cm$[228782];k1+=f9Cm$[481343];return table instanceof $[k1][f9Cm$.m96][s5A]?table:$(table)[Y7b]();};var _dtHighlight=function(node){node=$(node);I8Z.m_d();setTimeout(function(){var H$2="hig";var Y3O="hlight";var a_=H$2;a_+=Y3O;var G3=g66;G3+=E86;node[G3](a_);setTimeout(function(){var z6W='noHighlight';var N1h=550;var A3Z="highlig";var o9=A3Z;o9+=s$2;o9+=f9Cm$.J4L;node[k1$](z6W)[B_c](o9);setTimeout(function(){var A1n="removeCl";var L8M="lig";var C3n="noHigh";var b_=C3n;b_+=L8M;b_+=X$r;var V$=A1n;V$+=R77;I8Z.j$H();V$+=f9Cm$.J3V;node[V$](b_);},N1h);},J0Z);},c5X);};var _dtRowSelector=function(out,dt,identifier,fields,idFn){var r2=f9Cm$.t_T;r2+=f9Cm$.e08;r2+=q9Z;r2+=s$2;var Y6=f9Cm$.l60;Y6+=f9Cm$[23424];I8Z.m_d();Y6+=H3b;Y6+=f9Cm$.J3V;dt[Y6](identifier)[f_s]()[r2](function(idx){var H9Q='Unable to find row identifier';var b99=14;var G_=f9Cm$.l60;G_+=f9Cm$[23424];I8Z.j$H();G_+=H3b;var W2=f9Cm$.l60;W2+=f9Cm$[23424];W2+=H3b;var row=dt[W2](idx);var data=row[I$4]();var idSrc=idFn(data);if(idSrc === undefined){Editor[h8G](H9Q,b99);}out[idSrc]={data:data,fields:fields,idSrc:idSrc,node:row[z$j](),type:G_};});};var _dtFieldsFromIdx=function(dt,fields,idx,ignoreUnknown){var n_g=11;var v6M="ce. Please specify the field name.";var A1J="mData";var P3S="aoColumns";var U6E="editFi";var M$I="Unable to automatically determine field from sour";var V6v="itFi";var K1=f9Cm$.t_T;K1+=f9Cm$.e08;K1+=j1L;var N1=w3$;N1+=V6v;N1+=G6S;N1+=f9Cm$[555616];var B8=U6E;B8+=G6S;B8+=f9Cm$[555616];var H0=Y1m;H0+=I4O;H0+=K18;H0+=h7Q;var col=dt[H0]()[C37][P3S][idx];var dataSrc=col[B8] !== undefined?col[N1]:col[A1J];var resolvedFields={};var run=function(field,dataSrcIn){var h6=f9Cm$[481343];h6+=f9Cm$.e08;h6+=D6p;h6+=f9Cm$.t_T;if(field[h6]() === dataSrcIn){var L9=N1z;L9+=f9Cm$.t_T;resolvedFields[field[L9]()]=field;}};$[K1](fields,function(name,fieldInst){var d9=K_b;d9+=t$w;if(Array[d9](dataSrc)){for(var _i=C37,dataSrc_1=dataSrc;_i < dataSrc_1[B97];_i++){var data=dataSrc_1[_i];run(fieldInst,data);}}else {run(fieldInst,dataSrc);}});if($[E8u](resolvedFields) && !ignoreUnknown){var X3=M$I;X3+=v6M;Editor[h8G](X3,n_g);}return resolvedFields;};var _dtCellSelector=function(out,dt,identifier,allFields,idFn,forceFields){if(forceFields === void C37){forceFields=B3c;}var cells=dt[q5W](identifier);cells[f_s]()[a8N](function(idx){var q5H="splayF";var V48="Fi";var n4B="fixedNode";var P$M="playFields";var n$H="attachFields";var y8=k_d;y8+=f9Cm$[481343];y8+=h7H;y8+=K5v;var A9=e3G;A9+=f9Cm$.t_T;A9+=g5f;A9+=f9Cm$.J3V;var n$=q9Z;n$+=y7x;n$+=l1E;var B_=f9Cm$[555616];B_+=q94;var cell=dt[Z_2](idx);var row=dt[P2R](idx[P2R]);var data=row[B_]();var idSrc=idFn(data);var fields=forceFields || _dtFieldsFromIdx(dt,allFields,idx[p9_],cells[n$]() > Y5Y);var isNode=typeof identifier === f9Cm$.c8L && identifier[Y9y] || identifier instanceof $;var prevDisplayFields;var prevAttach;var prevAttachFields;if(Object[A9](fields)[y8]){var l8=C1Z;l8+=P$M;var k9=f9Cm$[481343];k9+=f9Cm$[23424];k9+=f9Cm$[555616];k9+=f9Cm$.t_T;var K0=O1K;K0+=s$2;var c1=O1K;c1+=R3u;var M6=f9Cm$.l60;M6+=f9Cm$[23424];M6+=H3b;if(out[idSrc]){var R6=h44;R6+=q5H;R6+=e6e;var h3=O1K;h3+=s$2;h3+=V48;h3+=K5X;var n5=f9Cm$.e08;n5+=f9Cm$.J4L;n5+=Q3e;n5+=j1L;prevAttach=out[idSrc][n5];prevAttachFields=out[idSrc][h3];prevDisplayFields=out[idSrc][R6];}_dtRowSelector(out,dt,idx[M6],allFields,idFn);out[idSrc][n$H]=prevAttachFields || [];out[idSrc][c1][Z_J](Object[b66](fields));out[idSrc][P8n]=prevAttach || [];out[idSrc][K0][Z_J](isNode?$(identifier)[W7y](C37):cell[n4B]?cell[n4B]():cell[k9]());out[idSrc][l8]=prevDisplayFields || ({});$[d7H](out[idSrc][p6K],fields);}});};var _dtColumnSelector=function(out,dt,identifier,fields,idFn){var G1=R27;G1+=f9Cm$.t_T;I8Z.m_d();G1+=f9Cm$.J3V;dt[q5W](B3c,identifier)[G1]()[a8N](function(idx){I8Z.j$H();_dtCellSelector(out,dt,idx,fields,idFn);});};var dataSource$1={commit:function(action,identifier,data,store){var w8S="earchPanes";var E8I="rch";var f2w="rebuild";var P8G="hBuilder";var b26="rowI";var t9b="onsive";var R9j="uil";var p6z="reb";var V9M="rchBuild";var B9y="wType";var h92="arc";var O5j="Bu";var p_Z="archB";var B45="sea";var W8p="alc";var z7J="ilder";var r$U="any";var K9Q="responsive";var l_l="rebuildPane";var s_R="nes";var e1O="getDetails";var I0n="raw";var J$Q="verSi";var Q0v="searchPa";var D$v="ui";var B5=f9Cm$[555616];B5+=f9Cm$.l60;B5+=f9Cm$.e08;B5+=B9y;var s6=P2R;s6+=h3G;s6+=f9Cm$[555616];s6+=f9Cm$.J3V;var b6=b26;b6+=f9Cm$[555616];b6+=f9Cm$.J3V;var m9=r21;m9+=f9Cm$.l60;m9+=J$Q;m9+=R3S;var C4=f9Cm$.J4L;C4+=j$2;C4+=k_d;var that=this;var dt=_dtApi(this[f9Cm$.J3V][C4]);var ssp=dt[q$G]()[C37][u8X][m9];var ids=store[b6];if(!_dtIsSsp(dt,this) && action === Z75 && store[s6][B97]){var row=void C37;var compare=function(id){I8Z.j$H();return function(rowIdx,rowData,rowNode){var Y8=q9Z;Y8+=f9Cm$.e08;Y8+=Q6Q;I8Z.j$H();Y8+=Q6Q;var f6=K18;f6+=f9Cm$[555616];return id == dataSource$1[f6][Y8](that,rowData);};};for(var i=C37,ien=ids[B97];i < ien;i++){var b2=f9Cm$.e08;b2+=f9Cm$[481343];b2+=g5f;try{var p1=f9Cm$.l60;p1+=f9Cm$[23424];p1+=H3b;row=dt[p1](safeQueryId(ids[i]));}catch(e){row=dt;}if(!row[r$U]()){row=dt[P2R](compare(ids[i]));}if(row[b2]() && !ssp){row[h6J]();}}}var drawType=this[f9Cm$.J3V][x6t][B5];if(drawType !== n50){var y7=p6z;y7+=D$v;y7+=Q6Q;y7+=f9Cm$[555616];var A1=Y1m;A1+=h92;A1+=P8G;var o0=Y1m;o0+=p_Z;o0+=R9j;o0+=U7U;var r7=Q0v;r7+=s_R;var B3=G$b;B3+=B1d;B3+=t9b;var s3=f9Cm$[555616];s3+=f9Cm$.l60;s3+=f9Cm$.e08;s3+=H3b;var V9=e2b;V9+=K5v;var dtAny=dt;if(ssp && ids && ids[V9]){var y6=f9Cm$[555616];y6+=I0n;dt[X9O](y6,function(){var l7=Q6Q;l7+=q8V;l7+=K5v;for(var i=C37,ien=ids[l7];i < ien;i++){var L_=f9Cm$.e08;L_+=f9Cm$[481343];L_+=g5f;var row=dt[P2R](safeQueryId(ids[i]));if(row[L_]()){var O4=L6$;O4+=f9Cm$.t_T;_dtHighlight(row[O4]());}}});}dt[s3](drawType);if(dtAny[B3]){var U3=f9Cm$.l60;U3+=d6G;U3+=W8p;dtAny[K9Q][U3]();}if(typeof dtAny[r7] === t0t && !ssp){var u_=f9Cm$.J3V;u_+=w8S;dtAny[u_][l_l](undefined,X17);}if(dtAny[o0] !== undefined && typeof dtAny[A1][y7] === t0t && !ssp){var j0=B45;j0+=V9M;j0+=F7I;var s_=B45;s_+=E8I;s_+=O5j;s_+=z7J;dtAny[s_][f2w](dtAny[j0][e1O]());}}},create:function(fields,data){var n2=Q3e;n2+=J0n;n2+=Q6Q;n2+=f9Cm$.t_T;var dt=_dtApi(this[f9Cm$.J3V][n2]);I8Z.m_d();if(!_dtIsSsp(dt,this)){var e0=k14;e0+=R3S;var h1=f9Cm$.l60;h1+=f9Cm$[23424];h1+=H3b;var row=dt[h1][g66](data);_dtHighlight(row[e0]());}},edit:function(identifier,fields,data,store){var a0g="ll";var i1S="ny";var m2h="ditOpt";var U9=f9Cm$.t_T;U9+=m2h;U9+=f9Cm$.J3V;var B9=Q3e;B9+=R6r;B9+=f9Cm$.t_T;var that=this;var dt=_dtApi(this[f9Cm$.J3V][B9]);if(!_dtIsSsp(dt,this) || this[f9Cm$.J3V][U9][s5Z] === n50){var x4=f9Cm$.e08;x4+=i1S;var m0=g5R;m0+=g5f;var r$=e7k;r$+=a0g;var b1=K18;b1+=f9Cm$[555616];var rowId_1=dataSource$1[b1][r$](this,data);var row=void C37;try{row=dt[P2R](safeQueryId(rowId_1));}catch(e){row=dt;}if(!row[m0]()){var Z3=f9Cm$.l60;Z3+=h98;row=dt[Z3](function(rowIdx,rowData,rowNode){var w_=q9Z;w_+=M_7;w_+=Q6Q;I8Z.m_d();var q2=K18;q2+=f9Cm$[555616];return rowId_1 == dataSource$1[q2][w_](that,rowData);});}if(row[x4]()){var k8=f9Cm$[555616];k8+=f9Cm$.e08;k8+=f9Cm$.J4L;k8+=f9Cm$.e08;var F5=f9Cm$[555616];F5+=q94;var toSave=extend({},row[F5](),X17);toSave=extend(toSave,data,X17);row[k8](toSave);var idx=$[G_5](rowId_1,store[O3l]);store[O3l][O9l](idx,Y5Y);}else {row=dt[P2R][g66](data);}_dtHighlight(row[z$j]());}},fakeRow:function(insertPoint){var c1w="draw.dte";var w1s='<td>';var r2x="ao";var y2N="-creat";var j_9="eInline";var A6J="isible";var P2l="dte-inlineAdd\"";var e2d="olum";var u4C=':visible';var J2D="count";var W_J="Col";var Y53="<tr clas";var U6x="s=\"";var b7$="sClass";var f2f="mns";var Y_=a_T;Y_+=H3b;var N6=f9Cm$[228782];N6+=s4W;N6+=o$D;var k$=c1w;k$+=y2N;k$+=j_9;var Q3=q9Z;Q3+=e2d;Q3+=f9c;var C_=Y53;C_+=U6x;C_+=P2l;C_+=i$Z;var t_=f9Cm$.J4L;t_+=j$2;t_+=Q6Q;t_+=f9Cm$.t_T;var dt=_dtApi(this[f9Cm$.J3V][t_]);var tr=$(C_);var attachFields=[];var attach=[];var displayFields={};I8Z.j$H();var tbody=dt[z6u](undefined)[f9H]();for(var i=C37,ien=dt[Q3](u4C)[J2D]();i < ien;i++){var d1=Q6Q;d1+=f9Cm$.t_T;d1+=g7N;d1+=s$2;var i7=r2x;i7+=W_J;i7+=f9Cm$.Y87;i7+=f2f;var l6=f9Cm$[228782];l6+=e6e;var X6=X2s;X6+=A6J;var visIdx=dt[p9_](i + X6)[R27]();var td=$(w1s)[A8d](tr);var fields=_dtFieldsFromIdx(dt,this[f9Cm$.J3V][l6],visIdx,X17);var settings=dt[q$G]()[C37];var className=settings[i7][visIdx][b7$];if(className){var j7=g66;j7+=B7X;j7+=R77;j7+=f9Cm$.J3V;td[j7](className);}if(Object[b66](fields)[d1]){attachFields[Z_J](Object[b66](fields));attach[Z_J](td[C37]);$[d7H](displayFields,fields);}}var append=function(){var U6u="mp";var L_H="ndT";var h_r="recordsDis";var U5=I6h;U5+=p7O;U5+=L_H;U5+=f9Cm$[23424];var X9=P2$;X9+=X8A;var J0=f9Cm$.t_T;J0+=f9Cm$[481343];J0+=f9Cm$[555616];var J$=h_r;J$+=b5Q;J$+=g5f;var Y$=K18;Y$+=f9Cm$[481343];Y$+=f9Cm$[228782];Y$+=f9Cm$[23424];var M8=D6_;M8+=P9U;if(dt[M8][Y$]()[J$] === C37){var s8=f9Cm$.t_T;s8+=U6u;s8+=f9Cm$.J4L;s8+=g5f;$(tbody)[s8]();}var action=insertPoint === J0?X9:U5;tr[action](tbody);};this[G2h]=tr;append();dt[h8a](k$,function(){append();});return {0:{attach:attach,attachFields:attachFields,displayFields:displayFields,fields:this[f9Cm$.J3V][N6],type:Y_}};},fakeRowEnd:function(){var o8L="dtFakeRow";var P8Z='draw.dte-createInline';var R6O="rec";var A0t="play";var m$K="dsDis";var Q7z="__";var D8=R6O;D8+=A5h;D8+=m$K;D8+=A0t;var s1=B96;s1+=g2v;var q8=B1d;q8+=D2Y;var Z_=Q06;Z_+=I7r;Z_+=f9Cm$.t_T;var B6=Q7z;B6+=o8L;var u5=f9Cm$[23424];u5+=o3v;var dt=_dtApi(this[f9Cm$.J3V][z6u]);dt[u5](P8Z);this[B6][Z_]();this[G2h]=B3c;if(dt[q8][s1]()[D8] === C37){dt[u8I](h4R);}},fields:function(identifier){var j4J="Obj";var R52="isPlain";var T6D="col";var b5=R52;b5+=j4J;b5+=d6G;b5+=f9Cm$.J4L;I8Z.m_d();var e5=s4L;e5+=o$D;var idFn=dataGet(this[f9Cm$.J3V][c5v]);var dt=_dtApi(this[f9Cm$.J3V][z6u]);var fields=this[f9Cm$.J3V][e5];var out={};if($[b5](identifier) && (identifier[X9o] !== undefined || identifier[m$k] !== undefined || identifier[q5W] !== undefined)){var e2=T6D;e2+=f9Cm$.Y87;e2+=D6p;e2+=f9c;if(identifier[X9o] !== undefined){_dtRowSelector(out,dt,identifier[X9o],fields,idFn);}if(identifier[e2] !== undefined){var Z6=A6T;Z6+=k8K;Z6+=f9Cm$[481343];Z6+=f9Cm$.J3V;_dtColumnSelector(out,dt,identifier[Z6],fields,idFn);}if(identifier[q5W] !== undefined){_dtCellSelector(out,dt,identifier[q5W],fields,idFn);}}else {_dtRowSelector(out,dt,identifier,fields,idFn);}return out;},id:function(data){var Y1=K18;Y1+=r4b;I8Z.m_d();Y1+=f9Cm$.l60;Y1+=q9Z;var idFn=dataGet(this[f9Cm$.J3V][Y1]);return idFn(data);},individual:function(identifier,fieldNames){var N9=f9Cm$[228782];N9+=K18;N9+=f9Cm$.t_T;N9+=o$D;var T3=Q3e;T3+=J0n;T3+=Q6Q;T3+=f9Cm$.t_T;var idFn=dataGet(this[f9Cm$.J3V][c5v]);var dt=_dtApi(this[f9Cm$.J3V][T3]);var fields=this[f9Cm$.J3V][N9];var out={};var forceFields;if(fieldNames){var z4=K18;z4+=f6t;z4+=J_S;z4+=g5f;if(!Array[z4](fieldNames)){fieldNames=[fieldNames];}forceFields={};$[a8N](fieldNames,function(i,name){I8Z.j$H();forceFields[name]=fields[name];});}_dtCellSelector(out,dt,identifier,fields,idFn,forceFields);return out;},prep:function(action,identifier,submit,json,store){var x_S="elle";var _this=this;if(action === c2D){var f$=f9Cm$[555616];f$+=f9Cm$.e08;f$+=f9Cm$.J4L;f$+=f9Cm$.e08;store[O3l]=$[X5_](json[f$],function(row){var G4=q9Z;G4+=f9Cm$.e08;G4+=Q6Q;G4+=Q6Q;var A4=K18;I8Z.j$H();A4+=f9Cm$[555616];return dataSource$1[A4][G4](_this,row);});}if(action === Z75){var O6=f9Cm$[555616];O6+=q94;var cancelled_1=json[s00] || [];store[O3l]=$[X5_](submit[O6],function(val,key){var Y3=K18;Y3+=i0Z;Y3+=V9A;var R2=f9Cm$[326480];R2+=Q3e;return !$[E8u](submit[R2][key]) && $[Y3](key,cancelled_1) === -Y5Y?key:undefined;});}else if(action === m8Z){var T1=c18;T1+=x_S;T1+=f9Cm$[555616];store[s00]=json[T1] || [];}},refresh:function(){var W6v="reload";var b9=f9Cm$.J4L;b9+=j$2;b9+=k_d;var dt=_dtApi(this[f9Cm$.J3V][b9]);I8Z.m_d();dt[N7l][W6v](B3c,h4R);},remove:function(identifier,fields,store){I8Z.j$H();var v5n="every";var that=this;var dt=_dtApi(this[f9Cm$.J3V][z6u]);var cancelled=store[s00];if(cancelled[B97] === C37){var D4=a_T;D4+=D80;dt[D4](identifier)[h6J]();}else {var o4=f9Cm$.l60;o4+=f9Cm$[23424];o4+=H3b;o4+=f9Cm$.J3V;var indexes_1=[];dt[X9o](identifier)[v5n](function(){var A6=f9Cm$[555616];I8Z.j$H();A6+=f9Cm$.e08;A6+=Q3e;var id=dataSource$1[U7A][B7t](that,this[A6]());if($[G_5](id,cancelled) === -Y5Y){indexes_1[Z_J](this[R27]());}});dt[o4](indexes_1)[h6J]();}}};function _htmlId(identifier){var R1I="Could not find an elem";var M$y="key";var W_f=": ";var G0L='[data-editor-id="';var W8N=" or `id` of";var k26="ent with `data-editor-id`";var l3=Q6Q;l3+=n3_;l3+=h5h;I8Z.j$H();var U8=M$y;U8+=Q6Q;U8+=z_k;if(identifier === U8){return $(document);}var specific=$(G0L + identifier + D9g);if(specific[l3] === C37){specific=typeof identifier === t3X?$(safeQueryId(identifier)):$(identifier);}if(specific[B97] === C37){var K2=R1I;K2+=k26;K2+=W8N;K2+=W_f;throw new Error(K2 + identifier);}return specific;}function _htmlEl(identifier,name){var T06='[data-editor-field="';I8Z.m_d();var context=_htmlId(identifier);return $(T06 + name + D9g,context);}function _htmlEls(identifier,names){var K7=Q6Q;K7+=n3_;K7+=i1I;K7+=s$2;var out=$();for(var i=C37,ien=names[K7];i < ien;i++){var m6=f9Cm$.e08;m6+=f9Cm$[555616];m6+=f9Cm$[555616];out=out[m6](_htmlEl(identifier,names[i]));}return out;}function _htmlGet(identifier,dataSrc){I8Z.m_d();var W5D="ue]";var V$Y="[d";var J5$="ta-editor-val";var d4=s$2;d4+=s3K;var O0=V$Y;O0+=f9Cm$.e08;O0+=J5$;O0+=W5D;var el=_htmlEl(identifier,dataSrc);return el[e3Q](O0)[B97]?el[S9h](M_g):el[d4]();}function _htmlSet(identifier,fields,data){var q5=f9Cm$.t_T;q5+=f9Cm$.e08;q5+=q9Z;I8Z.m_d();q5+=s$2;$[q5](fields,function(name,field){var S5$="ilt";var s3I='[data-editor-value]';var C9j="ttr";var val=field[C$F](data);if(val !== undefined){var Q2=f9Cm$[228782];Q2+=S5$;Q2+=F7I;var el=_htmlEl(identifier,field[s3Y]());if(el[Q2](s3I)[B97]){var l0=f9Cm$.e08;l0+=C9j;el[l0](M_g,val);}else {var z0=s$2;z0+=s3K;el[a8N](function(){var j1b="dNo";I8Z.m_d();var J5U="firstChild";var L8m="removeChild";var o0s="chil";var L3=o0s;L3+=j1b;L3+=Q1L;while(this[L3][B97]){this[L8m](this[J5U]);}})[z0](val);}}});}var dataSource={create:function(fields,data){I8Z.m_d();if(data){var j2=q9Z;j2+=f9Cm$.e08;j2+=Q6Q;j2+=Q6Q;var id=dataSource[U7A][j2](this,data);try{var T2=Q6Q;T2+=q8V;T2+=K5v;if(_htmlId(id)[T2]){_htmlSet(id,fields,data);}}catch(e){;}}},edit:function(identifier,fields,data){I8Z.m_d();var id=dataSource[U7A][B7t](this,data) || z$2;_htmlSet(id,fields,data);},fields:function(identifier){var t_v="keyles";var W0=f9Cm$.l60;W0+=f9Cm$[23424];W0+=H3b;var Q_=p3M;Q_+=s$2;var out={};if(Array[d2_](identifier)){for(var i=C37,ien=identifier[B97];i < ien;i++){var t6=q9Z;t6+=f9Cm$.e08;t6+=Q6Q;t6+=Q6Q;var res=dataSource[P5P][t6](this,identifier[i]);out[identifier[i]]=res[identifier[i]];}return out;}var data={};var fields=this[f9Cm$.J3V][P5P];if(!identifier){var H3=t_v;H3+=f9Cm$.J3V;identifier=H3;}$[Q_](fields,function(name,field){I8Z.j$H();var u4p="valToData";var val=_htmlGet(identifier,field[s3Y]());field[u4p](data,val === B3c?undefined:val);});out[identifier]={data:data,fields:fields,idSrc:identifier,node:document,type:W0};return out;},id:function(data){var Q0=K18;Q0+=C$G;Q0+=q9Z;var idFn=dataGet(this[f9Cm$.J3V][Q0]);return idFn(data);},individual:function(identifier,fieldNames){var q0R="eyless";var I9L='editor-id';var x2Y='[data-editor-id]';var F8b="-editor";var P$w='Cannot automatically determine field name from data source';var I4H="dSelf";var i3d="-field";var F7=q4u;F7+=q9Z;F7+=s$2;var v8=Q6Q;v8+=d3h;var g9=K18;g9+=f6t;g9+=f9Cm$.l60;g9+=V9A;var attachEl;if(identifier instanceof $ || identifier[Y9y]){var u8=g5R;u8+=I4H;var C8=f9Cm$.e08;C8+=J2K;C8+=k7Z;attachEl=identifier;if(!fieldNames){var c2=f9Cm$[555616];c2+=q94;c2+=F8b;c2+=i3d;fieldNames=[$(identifier)[S9h](c2)];}var back=$[f9Cm$.E4X][j87]?C8:u8;identifier=$(identifier)[z89](x2Y)[back]()[I$4](I9L);}if(!identifier){var I0=e3G;I0+=q0R;identifier=I0;}if(fieldNames && !Array[g9](fieldNames)){fieldNames=[fieldNames];}if(!fieldNames || fieldNames[v8] === C37){throw new Error(P$w);}var out=dataSource[P5P][B7t](this,identifier);var fields=this[f9Cm$.J3V][P5P];var forceFields={};$[F7](fieldNames,function(i,name){I8Z.j$H();forceFields[name]=fields[name];});$[a8N](out,function(id,set){var r6C='cell';var r4=F3o;r4+=f9Cm$.t_T;r4+=e30;r4+=f9Cm$.J3V;var o6=x22;o6+=f9Cm$.e08;o6+=j1L;var B1=O1K;B1+=R3u;var O_=C18;O_+=f9Cm$.t_T;set[O_]=r6C;set[B1]=[fieldNames];set[o6]=attachEl?$(attachEl):_htmlEls(identifier,fieldNames)[f8d]();set[r4]=fields;set[p6K]=forceFields;});return out;},initField:function(cfg){var c55="[da";var p9b="ta-e";var y7C="tor-label=\"";var F1=S1i;F1+=J0n;F1+=G6S;var y4=y9e;y4+=M3E;var k4=f9Cm$[555616];k4+=k6$;k4+=f9Cm$.e08;var k0=c55;k0+=p9b;k0+=h44;k0+=y7C;var label=$(k0 + (cfg[k4] || cfg[h2d]) + y4);if(!cfg[F1] && label[B97]){var U0=X$r;U0+=J5n;cfg[e9d]=label[U0]();}},remove:function(identifier,fields){if(identifier !== z$2){_htmlId(identifier)[h6J]();}}};var classNames={actions:{create:m5,edit:z_,remove:z_H},body:{content:A0,wrapper:L4},bubble:{bg:Y2p,close:b1t,liner:G87,pointer:M7,table:x_,wrapper:u20},field:{'disabled':S0,'error':W9C,'input':f5g,'inputControl':z7_,'label':E0,'msg-error':W3Q,'msg-info':d2P,'msg-label':M5,'msg-message':v1x,'multiInfo':V5,'multiNoEdit':T3V,'multiRestore':w8,'multiValue':Y9J,'namePrefix':i9,'processing':D_H,'typePrefix':N7G,'wrapper':i$y},footer:{content:B6C,wrapper:T2r},form:{button:B$,buttonInternal:p0Q,buttons:Q2R,content:i_,error:e2o,info:O2m,tag:j_l,wrapper:C$},header:{content:L5G,title:{tag:B3c,class:j_l},wrapper:V0},inline:{buttons:m$,liner:m7q,wrapper:h_},processing:{active:H27,indicator:D_H},wrapper:P$V};var displayed$2=h4R;var cssBackgroundOpacity=Y5Y;var dom$1={background:$(T7)[C37],close:$(z$)[C37],content:B3c,wrapper:$(i71 + Y9 + x8 + V8)[C37]};function findAttachRow(editor,attach){var q2L="Table";var k7b="heade";var L45="hea";var P4=L45;P4+=f9Cm$[555616];var X_=z5S;X_+=B1d;X_+=K18;var V2=f9Cm$[326480];V2+=Q3e;V2+=q2L;var x6=f9Cm$[228782];x6+=f9Cm$[481343];var dt=new $[x6][V2][X_](editor[f9Cm$.J3V][z6u]);if(attach === P4){var Z9=k7b;Z9+=f9Cm$.l60;return dt[z6u](undefined)[Z9]();;}else if(editor[f9Cm$.J3V][Y8d] === c2D){var l5=f9Cm$.J4L;l5+=f9Cm$.e08;l5+=V5s;return dt[l5](undefined)[L2d]();}else {var E6=f9Cm$[481343];E6+=f9Cm$[23424];E6+=f9Cm$[555616];E6+=f9Cm$.t_T;var V3=a_T;V3+=H3b;return dt[V3](editor[f9Cm$.J3V][n9R])[E6]();}}function heightCalc$1(dte){var T$Q="windowPadding";var k9e='maxHeight';var h1y="Body";var S8P="terH";var l7x="eight";var s77="_Content";var I8C="iv.DTE_Footer";var g5h="Hei";var J9=f9Cm$[555616];J9+=f9Cm$[23424];J9+=D6p;var j6=q9Z;j6+=f9Cm$.J3V;j6+=f9Cm$.J3V;var x0=M61;x0+=o6H;var c8=O7w;c8+=h1y;c8+=s77;var C3=q9Z;C3+=f9Cm$[23424];C3+=f9Cm$[481343];C3+=f9Cm$[228782];var v6=t7M;v6+=h7H;v6+=X$r;var X7=y7x;X7+=S8P;X7+=l7x;var F6=f9Cm$[555616];F6+=I8C;var w2=b9d;w2+=g5h;w2+=H2$;var d8=M61;d8+=o6H;var header=$(m9y,dom$1[d8])[w2]();var footer=$(F6,dom$1[A2A])[X7]();var maxHeight=$(window)[v6]() - envelope[C3][T$Q] * X1m - header - footer;$(c8,dom$1[x0])[j6](k9e,maxHeight);return $(dte[J9][A2A])[t7b]();}function hide$2(dte,callback){var I5c="ffsetHeight";if(!callback){callback=function(){};}if(displayed$2){var p9=f9Cm$[23424];p9+=I5c;var n7=f9Cm$.e08;n7+=M$D;n7+=Y0o;n7+=f9Cm$.t_T;$(dom$1[P8l])[n7]({top:-(dom$1[P8l][p9] + Y7g)},p4_,function(){var K8b='normal';var f2T="backg";var f9=f9Cm$[228782];f9+=O7s;I8Z.j$H();f9+=F6K;f9+=g7i;var i4=f2T;i4+=f9Cm$.l60;i4+=y7x;i4+=T52;$([dom$1[A2A],dom$1[i4]])[f9](K8b,function(){I8Z.m_d();var d8Y="det";var i$=d8Y;i$+=o$n;i$+=s$2;$(this)[i$]();callback();});});displayed$2=h4R;}}function init$1(){var G7v="v.DTED_Envelope_Contai";var x8X="ckground";var c4=n0o;c4+=x8X;var g1=M61;g1+=p7O;g1+=f9Cm$.l60;var i1=h44;i1+=G7v;i1+=I$m;dom$1[P8l]=$(i1,dom$1[g1])[C37];cssBackgroundOpacity=$(dom$1[c4])[X5f](B5Z);}function show$2(dte,callback){var T7N="Envelope";var Z8Z="TED_E";var E8p="lick.";var J3O="click.DTED_";var k$N="bac";var e0N="fset";var Q5v="resize.DTED";var K$s="gr";var u4z="D_En";var i5t="opa";var J5r="resize.D";var K7H="yle";var K9a="rou";var i0Q="DTED_Envelope";var w2y="cli";var A1E="rappe";var E0g="uto";var s2h="ck.";var z6c="ck.DTED_Envelope";var t$f="marginLef";var u9K="opacity";var c8D="ckg";var K2M="offset";var i4o="bloc";var R2q="mal";var X2U='px';var J_5="fadeIn";var h2v="velope";var y$e="k.DTE";var f$x='0';var r6u="Width";var y6N="ound";var h1A="nvelope";var Z81="nor";var D7U="city";var V0T="offsetHeight";var r9=Q5v;r9+=j7g;r9+=T7N;var W5=f9Cm$[23424];W5+=f9Cm$[481343];var S7=J5r;S7+=Z8Z;S7+=h1A;var I1=J3O;I1+=T7N;var p3=f9Cm$[23424];p3+=f9Cm$[481343];var D$=q9Z;D$+=E8p;D$+=i0Q;var g2=f9Cm$[23424];g2+=f9Cm$[228782];g2+=f9Cm$[228782];var a8=M61;a8+=p7O;a8+=f9Cm$.l60;var a4=l_m;a4+=y$e;a4+=u4z;a4+=h2v;var s0=w2y;s0+=s2h;s0+=i0Q;var f3=H_E;f3+=f9Cm$[228782];var B2=n95;B2+=Z8Z;B2+=h1A;var C1=w2y;C1+=z6c;var B7=f9Cm$[23424];B7+=f9Cm$[228782];B7+=f9Cm$[228782];var N4=f9Cm$.J4L;N4+=K18;N4+=f9Cm$.J4L;N4+=k_d;var T8=f9Cm$.e08;T8+=E0g;var v4=f9Cm$.e08;v4+=B1d;v4+=G7H;var W1=J0n;W1+=L0k;var v1=D_8;v1+=f9Cm$[555616];if(!callback){callback=function(){};}$(p9T)[v1](dom$1[W1])[v4](dom$1[A2A]);dom$1[P8l][r_x][Z$p]=T8;if(!displayed$2){var z8=q9Z;z8+=a3Y;z8+=u7S;var b4=H3b;b4+=A1E;b4+=f9Cm$.l60;var k3=Z81;k3+=R2q;var C0=k$N;C0+=e3G;C0+=K$s;C0+=y6N;var P3=i4o;P3+=e3G;var K8=f9Cm$.J3V;K8+=f9Cm$.J4L;K8+=z7F;K8+=f9Cm$.t_T;var D0=i5t;D0+=D7U;var j1=n0o;j1+=c8D;j1+=K9a;j1+=T52;var J1=B1d;J1+=G1H;var I3=f9Cm$.J4L;I3+=f9Cm$[23424];I3+=B1d;var A5=f9Cm$.J3V;A5+=f9Cm$.J4L;A5+=g5f;A5+=k_d;var Z8=Y$v;Z8+=u7S;var l2=B1d;l2+=G1H;var E$=f9Cm$.J4L;E$+=J_q;var I2=f9Cm$.J4L;I2+=f9Cm$[23424];I2+=B1d;var M$=U5a;M$+=z7F;M$+=f9Cm$.t_T;var c0=B1d;c0+=G1H;var I$=t$f;I$+=f9Cm$.J4L;var y1=H3b;y1+=t3F;y1+=F7I;var s7=g6l;s7+=f9Cm$.J4L;s7+=s$2;var v_=U5a;v_+=K7H;var d0=H_E;d0+=e0N;d0+=r6u;var e4=x22;e4+=L7Y;var r6=q9Z;r6+=f9Cm$[23424];r6+=f9Cm$[481343];r6+=f9Cm$[228782];var n3=U5a;n3+=z7F;n3+=f9Cm$.t_T;var m_=H3b;m_+=q7Q;m_+=Q8$;m_+=f9Cm$.l60;var style=dom$1[m_][n3];style[u9K]=f$x;style[G7z]=M20;var height=heightCalc$1(dte);var targetRow=findAttachRow(dte,envelope[r6][e4]);var width=targetRow[d0];style[G7z]=n50;style[u9K]=T1p;dom$1[A2A][v_][s7]=width + X2U;dom$1[y1][r_x][I$]=-(width / X1m) + c0;dom$1[A2A][M$][I2]=$(targetRow)[K2M]()[E$] + targetRow[V0T] + l2;dom$1[Z8][A5][I3]=-Y5Y * height - c5X + J1;dom$1[j1][r_x][D0]=f$x;dom$1[e74][K8][G7z]=P3;$(dom$1[C0])[K5h]({opacity:cssBackgroundOpacity},k3);$(dom$1[b4])[J_5]();$(dom$1[z8])[K5h]({top:C37},p4_,callback);}$(dom$1[H9B])[S9h](N4,dte[B5C][H9B])[B7](C1)[h8a](B2,function(e){dte[H9B]();});$(dom$1[e74])[f3](s0)[h8a](a4,function(e){dte[e74]();});$(J$x,dom$1[a8])[g2](D$)[p3](I1,function(e){var K6R="D_Envelope_Content_";var W3=B0_;W3+=i2S;W3+=K6R;W3+=i28;if($(e[Y_v])[P6T](W3)){dte[e74]();}});$(window)[v93](S7)[W5](r9,function(){heightCalc$1(dte);});displayed$2=X17;}var envelope={close:function(dte,callback){I8Z.m_d();hide$2(dte,callback);},conf:{attach:N7,windowPadding:Y7g},destroy:function(dte){hide$2();},init:function(dte){init$1();return envelope;},node:function(dte){var l9=H3b;l9+=m0O;return dom$1[l9][C37];},open:function(dte,append,callback){var H$N="ppendChild";var J3U="dC";var r2R="ontent";var r0_="hild";var l_=q9Z;l_+=h_t;var t$=f9Cm$.e08;t$+=w1H;t$+=J3U;t$+=r0_;var g3=q9Z;g3+=Y1k;g3+=f9Cm$.J4L;var B0=f9Cm$.e08;B0+=H$N;var M9=q9Z;M9+=r2R;var g8=s_E;g8+=j1L;var Z5=q9Z;Z5+=r0_;Z5+=f9Cm$.l60;Z5+=n3_;$(dom$1[P8l])[Z5]()[g8]();dom$1[M9][B0](append);dom$1[g3][t$](dom$1[l_]);show$2(dte,callback);}};function isMobile(){var K2d=576;var k0z="Wid";var k09="undefin";var i6S="orientation";var L2=b9d;L2+=k0z;L2+=K5v;var l4=k09;l4+=w3$;I8Z.j$H();return typeof window[i6S] !== l4 && window[L2] <= K2d?X17:h4R;}var displayed$1=h4R;var ready=h4R;var scrollTop=C37;var dom={background:$(N30),close:$(z9),content:B3c,wrapper:$(b40 + P8 + q1 + p4k + G6 + K3 + Q4m + Q4m)};function heightCalc(){var C4U="lc(10";var x17="px";var k1L="h - ";var m47="iv.DTE_";var J4T="indowPaddin";var x6f='div.DTE_Body_Content';var I8E="maxHeigh";var M6W="nf";var j5t="0v";var m6V="apper";var F1o="Fo";var h5V="He";var P2S="xHei";var O31="Body_Conte";var q9=O7w;q9+=F1o;q9+=e0Y;q9+=F7I;var e3=y7x;e3+=A7I;e3+=h5V;e3+=F4F;var m8=p5y;m8+=p26;m8+=F7I;var headerFooter=$(m9y,dom[m8])[e3]() + $(q9,dom[A2A])[t7b]();I8Z.m_d();if(isMobile()){var U7=x17;U7+=h8R;var G9=e7k;G9+=C4U;G9+=j5t;G9+=k1L;var W7=I8E;W7+=f9Cm$.J4L;var P9=p5y;P9+=m6V;var s$=f9Cm$[555616];s$+=m47;s$+=O31;s$+=l1E;$(s$,dom[P9])[X5f](W7,G9 + headerFooter + U7);}else {var B4=D6p;B4+=f9Cm$.e08;B4+=P2S;B4+=H2$;var j4=H3b;j4+=J4T;j4+=h7H;var w9=A6T;w9+=M6W;var maxHeight=$(window)[Z$p]() - self[w9][j4] * X1m - headerFooter;$(x6f,dom[A2A])[X5f](B4,maxHeight);}}function hide$1(dte,callback){var K0F="scro";var O$p="resize.DTED_Lightb";var Q6P="ox";var t9T="ackgr";var R30="llT";var z6F="oun";var N5=O$p;N5+=Q6P;var f_=J0n;f_+=t9T;f_+=z6F;f_+=f9Cm$[555616];var d7=q9Z;d7+=f9Cm$[23424];d7+=f9Cm$[481343];d7+=f9Cm$[228782];var n6=j7g;n6+=g5R;n6+=K18;n6+=r6r;var t9=K0F;t9+=R30;t9+=f9Cm$[23424];t9+=B1d;if(!callback){callback=function(){};}$(p9T)[t9](scrollTop);dte[n6](dom[A2A],{opacity:C37,top:self[d7][O9G]},function(){var Z0E="detac";var u9=Z0E;u9+=s$2;$(this)[u9]();callback();});dte[T4f](dom[f_],{opacity:C37},function(){I8Z.j$H();var Q$=f9Cm$[555616];Q$+=C6V;Q$+=L7Y;$(this)[Q$]();});displayed$1=h4R;$(window)[v93](N5);}function init(){var E77="pacit";var e2P='div.DTED_Lightbox_Content';var p2=f9Cm$[23424];p2+=E77;p2+=g5f;var n4=q9Z;n4+=f9Cm$.J3V;n4+=f9Cm$.J3V;var t2=H3b;t2+=t3F;t2+=F7I;if(ready){return;}dom[P8l]=$(e2P,dom[A2A]);dom[t2][n4](B5Z,C37);dom[e74][X5f](p2,C37);ready=X17;}function show$1(dte,callback){var G_w="click.DTED_Light";var m4T="aut";var J$9="llTop";var k4N='resize.DTED_Lightbox';var g1d="sc";var A9s="_anim";var P9z='click.DTED_Lightbox';var C3J="DTED_Ligh";var X52="TED_";var j6I="click.DTED_Li";var G83="box";var e_U="kgroun";var S4e="Lightbo";var z2s="tbox_Mobi";var F_=f9Cm$[23424];F_+=f9Cm$[481343];var s2=G_w;s2+=G83;var a$=f9Cm$[23424];a$+=o3v;var r5=j6I;r5+=H2$;r5+=r0K;r5+=G1H;var L1=f9Cm$[23424];L1+=f9Cm$[481343];var X0=f9Cm$[23424];X0+=f9Cm$[228782];X0+=f9Cm$[228782];var o_=n95;o_+=X52;o_+=S4e;o_+=G1H;var I_=f9Cm$[23424];I_+=f9Cm$[481343];var W9=M_R;W9+=J_E;var N_=p5y;N_+=f9Cm$.e08;N_+=y7w;var Z0=n0o;Z0+=q9Z;Z0+=e_U;Z0+=f9Cm$[555616];var y3=p26;y3+=n3_;y3+=f9Cm$[555616];if(isMobile()){var D9=C3J;D9+=z2s;D9+=k_d;$(p9T)[k1$](D9);}$(p9T)[y3](dom[Z0])[P2$](dom[N_]);heightCalc();if(!displayed$1){var V6=g1d;V6+=a_T;V6+=J$9;var n1=A9s;n1+=o2e;var H1=H3b;H1+=m0O;var v7=O1r;v7+=f9Cm$[481343];v7+=K18;v7+=r6r;var q3=E57;q3+=f9Cm$[228782];var T0=q9Z;T0+=i3q;var f5=m4T;f5+=f9Cm$[23424];var j3=t7M;j3+=H2$;var R5=q9Z;R5+=i3q;displayed$1=X17;dom[P8l][R5](j3,f5);dom[A2A][T0]({top:-self[q3][O9G]});dte[v7](dom[H1],{opacity:Y5Y,top:C37},callback);dte[n1](dom[e74],{opacity:Y5Y});$(window)[h8a](k4N,function(){I8Z.j$H();heightCalc();});scrollTop=$(p9T)[V6]();}dom[W9][S9h](f7V,dte[B5C][H9B])[v93](P9z)[I_](o_,function(e){var e6=q9Z;e6+=F7s;e6+=f9Cm$.J3V;e6+=f9Cm$.t_T;dte[e6]();});dom[e74][X0](P9z)[L1](r5,function(e){var N7k="stopImmediateP";I8Z.j$H();var u2h="pagatio";var G$=N7k;G$+=a_T;G$+=u2h;G$+=f9Cm$[481343];e[G$]();dte[e74]();});$(J$x,dom[A2A])[a$](s2)[F_](P9z,function(e){var h62="DTED_Lightbox_Co";var G97="ntent_";var H4=h62;H4+=G97;I8Z.j$H();H4+=i28;if($(e[Y_v])[P6T](H4)){var L$=J0n;L$+=L0k;e[T7J]();dte[L$]();}});}var self={close:function(dte,callback){I8Z.j$H();hide$1(dte,callback);},conf:{offsetAni:z6r,windowPadding:z6r},destroy:function(dte){I8Z.j$H();if(displayed$1){hide$1(dte);}},init:function(dte){init();return self;},node:function(dte){return dom[A2A][C37];},open:function(dte,append,callback){var p$_="conte";var p8=q9Z;p8+=C2Q;p8+=f9Cm$.t_T;var v0=C$g;v0+=B1d;v0+=f9Cm$.t_T;v0+=T52;var E4=R3S;E4+=f9Cm$.J4L;E4+=o$n;E4+=s$2;var O$=p$_;O$+=l1E;var content=dom[O$];content[n_U]()[E4]();content[v0](append)[P2$](dom[p8]);show$1(dte,callback);}};var DataTable$5=$[z7][f9Cm$.m96];function add(cfg,after,reorder){var O5Y="splayReorder";var E2t="nshi";var m7_="inA";var u12='Error adding field \'';var N2G="reverse";var U1z="itField";var E09='Error adding field. The field requires a `name` option';var A2o="isplayReorder";var A68=" A fie";var g52="_di";var i8e="\'.";var f4E="ld already exists with this name";var x8z="Res";var q2W="sh";var m2=F3o;m2+=f9Cm$.t_T;m2+=o$D;var I8=D6p;I8+=f9Cm$[23424];I8+=f9Cm$[555616];I8+=f9Cm$.t_T;I8Z.j$H();var U2=y7j;U2+=K18;U2+=f9Cm$.t_T;U2+=e30;var d$=B96;d$+=U1z;var t4=f9Cm$[228782];t4+=K18;t4+=D9y;t4+=f9Cm$.J3V;var K$=K_b;K$+=z5S;K$+=J_S;K$+=g5f;if(reorder === void C37){reorder=X17;}if(Array[K$](cfg)){var G5=g52;G5+=O5Y;var H2=Q6Q;H2+=f9Cm$.t_T;H2+=L47;H2+=K5v;if(after !== undefined){cfg[N2G]();}for(var _i=C37,cfg_1=cfg;_i < cfg_1[H2];_i++){var cfgDp=cfg_1[_i];this[g66](cfgDp,after,h4R);}this[G5](this[U2b]());return this;}var name=cfg[h2d];if(name === undefined){throw new Error(E09);}if(this[f9Cm$.J3V][t4][name]){var T6=i8e;T6+=A68;T6+=f4E;throw new Error(u12 + name + T6);}this[e3V](d$,cfg);var editorField=new Editor[U2](cfg,this[A2x][D_P],this);if(this[f9Cm$.J3V][I8]){var w0=C95;w0+=x8z;w0+=C6V;var X2=o$O;X2+=o$D;var editFields=this[f9Cm$.J3V][X2];editorField[w0]();$[a8N](editFields,function(idSrc,editIn){var k3b="valFro";var F$9="mD";var Q_C="iS";var h$=U3L;h$+=Q_C;h$+=C6V;var Y5=f9Cm$[555616];Y5+=f9Cm$.e08;Y5+=f9Cm$.J4L;Y5+=f9Cm$.e08;var value;if(editIn[Y5]){var i3=f9Cm$[555616];i3+=k6$;i3+=f9Cm$.e08;var f4=k3b;f4+=F$9;f4+=q94;value=editorField[f4](editIn[i3]);}editorField[h$](idSrc,value !== undefined?value:editorField[Z87]());});}this[f9Cm$.J3V][m2][name]=editorField;if(after === undefined){var w3=B1d;w3+=f9Cm$.Y87;w3+=q2W;this[f9Cm$.J3V][U2b][w3](name);}else if(after === B3c){var O2=f9Cm$.Y87;O2+=E2t;O2+=u3w;var A_=A5h;A_+=R3S;A_+=f9Cm$.l60;this[f9Cm$.J3V][A_][O2](name);}else {var C6=G5D;C6+=f9Cm$.t_T;C6+=f9Cm$.l60;var U_=m7_;U_+=C58;U_+=n7x;var idx=$[U_](after,this[f9Cm$.J3V][U2b]);this[f9Cm$.J3V][C6][O9l](idx + Y5Y,C37,name);}if(reorder !== h4R){var Y4=I_j;Y4+=A2o;this[Y4](this[U2b]());}return this;}function ajax(newAjax){var h7=P9q;h7+=J_Y;if(newAjax){this[f9Cm$.J3V][N7l]=newAjax;return this;}return this[f9Cm$.J3V][h7];}function background(){var A77="onBackground";var n4d='blur';var p$=f9Cm$[228782];p$+=s3_;p$+=h8a;var onBackground=this[f9Cm$.J3V][x6t][A77];if(typeof onBackground === p$){onBackground(this);}else if(onBackground === n4d){var E3=J0n;E3+=Q6Q;E3+=f9Cm$.Y87;E3+=f9Cm$.l60;this[E3]();}else if(onBackground === K0N){var g5=q9Z;g5+=F7s;g5+=f9Cm$.J3V;g5+=f9Cm$.t_T;this[g5]();}else if(onBackground === O2n){this[n3l]();}return this;}function blur(){var R1=j7g;R1+=J0n;R1+=c5p;I8Z.m_d();R1+=f9Cm$.l60;this[R1]();return this;}function bubble(cells,fieldNames,showIn,opts){var b0E="taSou";var F4Y="ainObje";var R9=j7g;R9+=w3$;R9+=K18;R9+=f9Cm$.J4L;var E7=I_j;E7+=f9Cm$.e08;E7+=b0E;E7+=j21;var v$=o27;v$+=F4Y;v$+=Q2A;var p5=b7w;p5+=K18;p5+=f9Cm$[555616];p5+=g5f;var _this=this;if(showIn === void C37){showIn=X17;}var that=this;if(this[p5](function(){I8Z.j$H();var S4=p0q;S4+=J0n;S4+=Q6Q;S4+=f9Cm$.t_T;that[S4](cells,fieldNames,opts);})){return this;}if($[v$](fieldNames)){opts=fieldNames;fieldNames=undefined;showIn=X17;}else if(typeof fieldNames === I06){showIn=fieldNames;fieldNames=undefined;opts=undefined;}if($[F7v](showIn)){opts=showIn;showIn=X17;}if(showIn === undefined){showIn=X17;}opts=$[d7H]({},this[f9Cm$.J3V][U4K][m0D],opts);var editFields=this[E7](k8T,cells,fieldNames);this[R9](cells,editFields,t0m,opts,function(){var m4X="ze.";var l26="/div";var U5Q='<div class="DTE_Processing_Indicator"><span></div>';var V03="pointer";var K0_="concat";var d_y="><div></div></div>";var a3E="liner";var r4_="</di";var B27="nimate";var A69="_post";var B6t="he";var A98="las";var j5e="leNodes";var s8T="\" title=";var D6x="resi";var O4U="child";var I19="bblePosition";var h_$=' scroll.';var A7M="prepe";var u3=O1r;u3+=B27;var g6=A69;g6+=J_q;g6+=f9Cm$.t_T;g6+=f9Cm$[481343];var v5=q1i;v5+=I19;var R3=f9Cm$[23424];R3+=f9Cm$[481343];var s9=f9Cm$[23424];s9+=f9Cm$[481343];var A3=O1b;A3+=f9Cm$[555616];var t7=f9Cm$.e08;t7+=f9Cm$[555616];t7+=f9Cm$[555616];var M4=J0n;M4+=f9Cm$.Y87;M4+=C_v;M4+=f9c;var q_=B4Y;q_+=p0O;q_+=f9Cm$.t_T;var V4=f9Cm$[555616];V4+=f9Cm$[23424];V4+=D6p;var O8=A7M;O8+=f9Cm$[481343];O8+=f9Cm$[555616];var p_=C$g;p_+=B1d;p_+=f9Cm$.t_T;p_+=T52;var a2=j1L;a2+=K18;a2+=e30;a2+=w2Q;var Y0=f9Cm$.t_T;Y0+=R2$;var G8=O4U;G8+=O4Y;G8+=f9Cm$[481343];var r0=r4_;r0+=T5I;var J4=s4A;J4+=j0k;J4+=T5I;var R7=k8A;R7+=l26;R7+=i$Z;var m7=u_6;m7+=O$n;m7+=i$Z;var K5=K18;K5+=W8E;K5+=t1_;var h5=s8T;h5+=y9e;var f7=q9Z;f7+=Q6Q;f7+=i$r;f7+=f9Cm$.t_T;var a0=r1X;a0+=v19;a0+=y9e;var g$=M61;g$+=p7O;g$+=f9Cm$.l60;var q6=y9e;q6+=d_y;var R$=J0n;R$+=h7H;var x3=J0n;x3+=f9Cm$.Y87;x3+=B5f;x3+=k_d;var w4=q9Z;w4+=A98;w4+=P_W;var R8=k6$;R8+=f9Cm$.J4L;R8+=o$n;R8+=s$2;var f2=p0q;f2+=J0n;f2+=j5e;var a7=D6x;a7+=m4X;var J5=f9Cm$[23424];J5+=f9Cm$[481343];var namespace=_this[Q49](opts);var ret=_this[w6o](t0m);if(!ret){return _this;}$(window)[J5](a7 + namespace + h_$ + namespace,function(){var L9j="bubblePosition";_this[L9j]();});var nodes=[];_this[f9Cm$.J3V][f2]=nodes[K0_][D08](nodes,pluck(editFields,R8));var classes=_this[w4][x3];var backgroundNode=$(v4Q + classes[R$] + q6);var container=$(v4Q + classes[g$] + x$l + v4Q + classes[a3E] + x$l + a0 + classes[z6u] + x$l + v4Q + classes[f7] + h5 + _this[K5][H9B] + m7 + U5Q + R7 + Q4m + v4Q + classes[V03] + J4 + r0);if(showIn){var m4=R$W;m4+=g5f;var t5=p26;t5+=n3_;t5+=f9Cm$[555616];t5+=X8A;container[t5](m4);backgroundNode[A8d](p9T);}var liner=container[G8]()[Y0](C37);var tableNode=liner[a2]();var closeNode=tableNode[n_U]();liner[p_](_this[X6t][c2Q]);tableNode[O8](_this[V4][O94]);if(opts[c1_]){var F4=O94;F4+=Y5K;F4+=f9Cm$[23424];var y5=f9Cm$[555616];y5+=f9Cm$[23424];y5+=D6p;liner[g9P](_this[y5][F4]);}if(opts[q_]){var U1=B6t;U1+=O7s;U1+=f9Cm$.l60;var y0=f9Cm$[555616];y0+=f9Cm$[23424];y0+=D6p;var T9=I6h;T9+=B1d;T9+=f9Cm$.t_T;T9+=T52;liner[T9](_this[y0][U1]);}if(opts[M4]){var b0=q1i;b0+=I4O;b0+=f9Cm$[23424];b0+=f9c;var z3=C$g;z3+=V7f;z3+=f9Cm$[555616];tableNode[z3](_this[X6t][b0]);}var finish=function(){var c2i="rDynamicInfo";var a2J="_clea";var r1=p3g;I8Z.m_d();r1+=Q$x;r1+=f9Cm$.J4L;var e9=a2J;e9+=c2i;_this[e9]();_this[r1](u7C,[t0m]);};var pair=$()[t7](container)[A3](backgroundNode);_this[k7m](function(submitComplete){_this[T4f](pair,{opacity:C37},function(){var I1v='resize.';I8Z.m_d();if(this === container[C37]){pair[Z3_]();$(window)[v93](I1v + namespace + h_$ + namespace);finish();}});});backgroundNode[s9](F7k,function(){I8Z.m_d();_this[h1_]();});closeNode[R3](F7k,function(){var J6=h7m;I8Z.j$H();J6+=Y1m;_this[J6]();});_this[v5]();_this[g6](t0m,h4R);var opened=function(){var a$r="ubbl";var q7=J0n;q7+=a$r;q7+=f9Cm$.t_T;var L6=f9Cm$[228782];L6+=J8D;L6+=M42;_this[U5L](_this[f9Cm$.J3V][u15],opts[L6]);_this[G29](I27,[q7,_this[f9Cm$.J3V][Y8d]]);};_this[u3](pair,{opacity:Y5Y},function(){I8Z.m_d();if(this === container[C37]){opened();}});});return this;}function bubblePosition(){var q_d="v.D";var I2i="scrollTop";var s30="Nodes";var U8k="eft";var c$N="le_Liner";var u7V="bottom";var n43="lef";var U_h="leBottom";var l0l="Clas";var u4d='top';var R2b="TE_Bubb";var w49="idt";var t8r="rW";var f6V='below';var K4I="rH";var a19="bubbleBottom";var Y3Y="eig";var O67='div.DTE_Bubble';var d$w="right";var c9m="oute";var C6v="nerHe";var W3A="bott";var h8=H$0;h8+=B1d;var o3=k_d;o3+=f9Cm$[481343];o3+=h7H;o3+=K5v;var I5=B96;I5+=C6v;I5+=F4F;var E9=f9Cm$.J4L;E9+=J_q;var k5=D3W;k5+=h5h;var o7=N8o;o7+=g8Y;o7+=X3e;var H_=q9Z;H_+=i3q;var u0=c9m;u0+=K4I;u0+=Y3Y;u0+=X$r;var H5=c9m;H5+=t8r;H5+=w49;H5+=s$2;var M2=Q6Q;M2+=U8k;var u2=f9Cm$.J4L;u2+=f9Cm$[23424];u2+=B1d;var K6=W3A;K6+=u8G;var O7=k_d;O7+=y8m;var M0=f9Cm$.l60;M0+=K18;M0+=h7H;M0+=X$r;var p7=D3W;p7+=h7H;p7+=K5v;var W$=n43;W$+=f9Cm$.J4L;var P$=f9Cm$.t_T;P$+=f9Cm$.e08;P$+=q9Z;P$+=s$2;var s5=m0D;s5+=s30;var l$=h44;l$+=q_d;l$+=R2b;l$+=c$N;var wrapper=$(O67);var liner=$(l$);var nodes=this[f9Cm$.J3V][s5];var position={bottom:C37,left:C37,right:C37,top:C37};$[P$](nodes,function(i,nodeIn){var D8C="offsetWidth";var B9g="ffsetHeig";var d_=f9Cm$[23424];d_+=B9g;d_+=s$2;d_+=f9Cm$.J4L;var i6=Q6Q;i6+=o2Y;i6+=f9Cm$.J4L;var c_=f9Cm$.l60;c_+=K18;c_+=h7H;c_+=X$r;var m3=Q6Q;m3+=f9Cm$.t_T;m3+=f9Cm$[228782];m3+=f9Cm$.J4L;var v9=k_d;v9+=u3w;var Q4=f9Cm$.J4L;Q4+=f9Cm$[23424];Q4+=B1d;var g7=f9Cm$.J4L;g7+=f9Cm$[23424];g7+=B1d;var O9=h7H;O9+=f9Cm$.t_T;O9+=f9Cm$.J4L;var c9=H_E;c9+=f9Cm$[228782];c9+=D0u;var pos=$(nodeIn)[c9]();nodeIn=$(nodeIn)[O9](C37);position[g7]+=pos[Q4];position[v9]+=pos[m3];position[c_]+=pos[i6] + nodeIn[D8C];position[u7V]+=pos[j52] + nodeIn[d_];});position[j52]/=nodes[B97];position[W$]/=nodes[p7];position[M0]/=nodes[O7];position[K6]/=nodes[B97];var top=position[u2];var left=(position[M2] + position[d$w]) / X1m;var width=liner[H5]();var height=liner[u0]();var visLeft=left - width / X1m;var visRight=visLeft + width;var docWidth=$(window)[T4U]();var viewportTop=$(window)[I2i]();var padding=T5p;wrapper[H_]({left:left,top:this[f9Cm$.J3V][a19]?position[u7V]:top});if(this[f9Cm$.J3V][a19]){var A$=J0n;A$+=f9Cm$.t_T;A$+=Q6Q;A$+=h98;wrapper[k1$](A$);}var curPosition=wrapper[o7]();if(liner[k5] && curPosition[E9] + height > viewportTop + window[I5]){var b7=h6J;b7+=l0l;b7+=f9Cm$.J3V;wrapper[X5f](u4d,top)[b7](f6V);this[f9Cm$.J3V][a19]=h4R;}else if(liner[o3] && curPosition[h8] - height < viewportTop){var z6=p0q;z6+=J0n;z6+=U_h;var a6=f9Cm$.e08;a6+=J2K;a6+=l0l;a6+=f9Cm$.J3V;var y2=f9Cm$.J4L;y2+=f9Cm$[23424];y2+=B1d;var u1=q9Z;u1+=f9Cm$.J3V;u1+=f9Cm$.J3V;wrapper[u1](y2,position[u7V])[a6](f6V);this[f9Cm$.J3V][z6]=X17;}if(visRight + padding > docWidth){var E5=Q6Q;E5+=U8k;var f0=q9Z;f0+=f9Cm$.J3V;f0+=f9Cm$.J3V;var diff=visRight - docWidth;liner[f0](E5,visLeft < padding?-(visLeft - padding):-(diff + padding));}else {var S2=Q6Q;S2+=U8k;var m1=q9Z;m1+=f9Cm$.J3V;m1+=f9Cm$.J3V;liner[m1](S2,visLeft < padding?-(visLeft - padding):C37);}return this;}function buttons(buttonsIn){var F6k="mpty";var o0E="18n";var I6=f9Cm$.t_T;I6+=f9Cm$.e08;I6+=q9Z;I6+=s$2;var f8=f9Cm$.t_T;f8+=F6k;var o8=f9Cm$[555616];o8+=f9Cm$[23424];o8+=D6p;var _this=this;if(buttonsIn === f9B){var p0=f9Cm$.J3V;p0+=q1I;p0+=f9Cm$.J4L;var r8=K18;r8+=o0E;buttonsIn=[{action:function(){this[n3l]();},text:this[r8][this[f9Cm$.J3V][Y8d]][p0]}];}else if(!Array[d2_](buttonsIn)){buttonsIn=[buttonsIn];}$(this[o8][d7J])[f8]();$[I6](buttonsIn,function(i,btn){var Q6H="></button";I8Z.m_d();var k2h="text";var u0G="tabInde";var k_G="keyu";var L4j="ndex";var s0e="bi";var B0i="<button";var R$R='keypress';var L9A="ssN";var N4R="pendT";var Q7=E2j;Q7+=f9Cm$.J4L;Q7+=f9Cm$[23424];Q7+=f9c;var G2=C$g;G2+=N4R;G2+=f9Cm$[23424];var C5=q9Z;C5+=k1J;C5+=K2u;var j$=f9Cm$[23424];j$+=f9Cm$[481343];var P5=k_G;P5+=B1d;var M_=f9Cm$[23424];M_+=f9Cm$[481343];var j8=J7A;j8+=h3G;j8+=T52;j8+=X2q;var x1=u0G;x1+=G1H;var L5=f9Cm$.J4L;L5+=f9Cm$.e08;L5+=s0e;L5+=L4j;var N2=f9Cm$[228782];N2+=s3_;N2+=h8a;var P1=s$2;P1+=f9Cm$.J4L;P1+=D6p;P1+=Q6Q;var D3=M_R;D3+=f9Cm$.e08;D3+=L9A;D3+=o1Y;var J3=a3D;J3+=f9Cm$[481343];var d6=B0i;d6+=Q6H;d6+=i$Z;var c5=f9Cm$.e08;c5+=q9Z;c5+=f9Cm$.J4L;c5+=X3e;var w5=S1i;w5+=k0d;if(typeof btn === t3X){btn={action:function(){this[n3l]();},text:btn};}var text=btn[k2h] || btn[w5];var action=btn[c5] || btn[f9Cm$.E4X];var attr=btn[S9h] || ({});$(d6,{class:_this[A2x][O94][J3] + (btn[D3]?E$d + btn[P03]:j_l)})[P1](typeof text === N2?text(_this):text || j_l)[S9h](L5,btn[x1] !== undefined?btn[j8]:C37)[S9h](attr)[M_](P5,function(e){I8Z.m_d();if(e[g4b] === m$Y && action){var d3=e7k;d3+=Q6Q;d3+=Q6Q;action[d3](_this);}})[h8a](R$R,function(e){var M5d="wh";var J0j="Defa";var G0=M5d;G0+=B1B;G0+=s$2;I8Z.m_d();if(e[G0] === m$Y){var h0=L5f;h0+=J0j;h0+=F8_;e[h0]();}})[j$](C5,function(e){e[G8l]();if(action){var n9=T93;n9+=Q6Q;action[n9](_this,e);}})[G2](_this[X6t][Q7]);});return this;}function clear(fieldName){var B0Q="_fie";var V1=q1F;V1+=f9Cm$[555616];V1+=f9Cm$.J3V;var that=this;I8Z.j$H();var sFields=this[f9Cm$.J3V][V1];if(typeof fieldName === t3X){var v3=B96;v3+=t$w;that[D_P](fieldName)[J8j]();delete sFields[fieldName];var orderIdx=$[v3](fieldName,this[f9Cm$.J3V][U2b]);this[f9Cm$.J3V][U2b][O9l](orderIdx,Y5Y);var includeIdx=$[G_5](fieldName,this[f9Cm$.J3V][u15]);if(includeIdx !== -Y5Y){this[f9Cm$.J3V][u15][O9l](includeIdx,Y5Y);}}else {var z1=B0Q;z1+=L_b;z1+=B12;var F0=p3M;F0+=s$2;$[F0](this[z1](fieldName),function(i,name){that[M1r](name);});}return this;}function close(){this[R4h](h4R);return this;}function create(arg1,arg2,arg3,arg4){var D3N="tionCl";var x3i="ber";var r_h="itC";var d$f="_even";var I_7="dif";var U0V="num";var J51="ier";var g1w="_crudArgs";var B_U="editField";var O1=B96;O1+=r_h;O1+=f9Cm$.l60;O1+=n_a;var N0=d$f;N0+=f9Cm$.J4L;var Q9=q4u;Q9+=j1L;var a1=F3o;a1+=f9Cm$.t_T;a1+=o$D;var Z4=I1O;Z4+=D3N;Z4+=S$6;var e$=h44;e$+=V1c;e$+=I8$;var D5=D6p;D5+=f9Cm$[23424];D5+=I_7;D5+=J51;var M3=f9Cm$.e08;M3+=q9Z;M3+=q5N;M3+=f9Cm$[481343];var V7=p9s;V7+=K18;V7+=f9Cm$[481343];var g4=o$O;g4+=o$D;var k6=U0V;k6+=x3i;var _this=this;var that=this;var sFields=this[f9Cm$.J3V][P5P];var count=Y5Y;if(this[M1B](function(){var t8=q9Z;t8+=f9Cm$.l60;t8+=q4u;t8+=U58;that[t8](arg1,arg2,arg3,arg4);})){return this;}if(typeof arg1 === k6){count=arg1;arg1=arg2;arg2=arg3;}this[f9Cm$.J3V][g4]={};for(var i=C37;i < count;i++){var n_=B_U;n_+=f9Cm$.J3V;this[f9Cm$.J3V][n_][i]={fields:this[f9Cm$.J3V][P5P]};}var argOpts=this[g1w](arg1,arg2,arg3,arg4);this[f9Cm$.J3V][h56]=V7;this[f9Cm$.J3V][M3]=c2D;this[f9Cm$.J3V][D5]=B3c;this[X6t][O94][r_x][e$]=M20;this[Z4]();this[Z4X](this[a1]());$[Q9](sFields,function(name,fieldIn){var D7=Y1m;I8Z.j$H();D7+=f9Cm$.J4L;var def=fieldIn[Z87]();fieldIn[e3n]();for(var i=C37;i < count;i++){fieldIn[Y$8](i,def);}fieldIn[D7](def);});this[N0](O1,B3c,function(){var b6t="may";var S6P="beO";var H$M="mOptions";var a$o="bleMain";var c7=b6t;c7+=S6P;c7+=p7O;c7+=f9Cm$[481343];var w6=j7g;w6+=f9Cm$[228782];w6+=A5h;w6+=H$M;var D2=w57;D2+=A1Z;D2+=a$o;_this[D2]();_this[w6](argOpts[w3u]);argOpts[c7]();});return this;}function undependent(parent){var k7g='.edep';var N$=f9Cm$[23424];N$+=f9Cm$[228782];N$+=f9Cm$[228782];var x$=f9Cm$[481343];x$+=x4Y;var A8=f9Cm$[228782];A8+=K18;A8+=G6S;A8+=f9Cm$[555616];if(Array[d2_](parent)){for(var i=C37,ien=parent[B97];i < ien;i++){this[p2n](parent[i]);}return this;}$(this[A8](parent)[x$]())[N$](k7g);return this;}function dependent(parent,url,optsIn){var b4l="POS";var W5X="pendent";var F9=I9t;F9+=f9Cm$[555616];F9+=P9K;var j9=f9Cm$.t_T;j9+=I7r;j9+=f9Cm$.t_T;j9+=l1E;var T5=f9Cm$[23424];T5+=f9Cm$[481343];var q$=f9Cm$[481343];q$+=f9Cm$[23424];q$+=f9Cm$[555616];q$+=f9Cm$.t_T;var f1=i9j;f1+=f9Cm$[555616];var t1=b4l;t1+=u08;var _this=this;if(Array[d2_](parent)){for(var i=C37,ien=parent[B97];i < ien;i++){var y$=R3S;y$+=W5X;this[y$](parent[i],url,optsIn);}return this;}I8Z.m_d();var that=this;var parentField=this[D_P](parent);var ajaxOpts={dataType:h48,type:t1};var opts=$[f1]({},{data:B3c,event:z_j,postUpdate:B3c,preUpdate:B3c},optsIn);var update=function(json){var C3E='show';var u$m="preUpdate";var S52='val';var R9b="pda";var e6Z="pd";var Q7E="postUpdate";var T8N='hide';var a2s='message';var m3r="ena";var h21="preU";var P2=m3r;P2+=R6r;P2+=f9Cm$.t_T;var C7=f9Cm$.t_T;C7+=o$n;C7+=s$2;var I4=f9Cm$.Y87;I4+=R9b;I4+=U58;var Q8=f9Cm$.t_T;Q8+=f9Cm$.l60;Q8+=f9Cm$.l60;Q8+=A5h;if(opts[u$m]){var w$=h21;w$+=e6Z;w$+=k6$;w$+=f9Cm$.t_T;opts[w$](json);}$[a8N]({errors:Q8,labels:B6L,messages:a2s,options:I4,values:S52},function(jsonProp,fieldFn){I8Z.j$H();if(json[jsonProp]){$[a8N](json[jsonProp],function(fieldIn,valIn){I8Z.m_d();that[D_P](fieldIn)[fieldFn](valIn);});}});$[C7]([T8N,C3E,P2,O6K],function(i,key){if(json[key]){var S6=f9Cm$.e08;S6+=M$D;S6+=Y0o;S6+=f9Cm$.t_T;that[key](json[key],json[S6]);}});if(opts[Q7E]){opts[Q7E](json);}parentField[T0W](h4R);};$(parentField[q$]())[T5](opts[j9] + F9,function(e){var t_k="editFiel";var v5N="values";var e3A=f9Cm$[555616];e3A+=f9Cm$.e08;e3A+=f9Cm$.J4L;e3A+=f9Cm$.e08;var c4B=I7r;c4B+=f9Cm$.e08;c4B+=Q6Q;var I05=f9Cm$.l60;I05+=f9Cm$[23424];I05+=D80;var H$=f9Cm$.l60;H$+=h98;var i8=t_k;i8+=f9Cm$[555616];i8+=f9Cm$.J3V;var j5=C6T;j5+=x6i;j5+=f9Cm$.J3V;var h9=f9Cm$.l60;h9+=f9Cm$[23424];h9+=D80;if($(parentField[z$j]())[I$n](e[Y_v])[B97] === C37){return;}parentField[T0W](X17);var data={};data[h9]=_this[f9Cm$.J3V][j5]?pluck(_this[f9Cm$.J3V][i8],v$r):B3c;I8Z.j$H();data[H$]=data[X9o]?data[I05][C37]:B3c;data[v5N]=_this[c4B]();if(opts[e3A]){var ret=opts[I$4](data);if(ret){data=ret;}}if(typeof url === t0t){var o=url[B7t](_this,parentField[P1N](),data,update,e);if(o){var x6r=f9Cm$[228782];x6r+=e7h;x6r+=O74;if(typeof o === f9Cm$.c8L && typeof o[y$I] === x6r){o[y$I](function(resolved){I8Z.j$H();if(resolved){update(resolved);}});}else {update(o);}}}else {var s5O=P9q;s5O+=f9Cm$.e08;s5O+=G1H;if($[F7v](url)){var y_q=X2q;y_q+=U58;y_q+=f9Cm$[481343];y_q+=f9Cm$[555616];$[y_q](ajaxOpts,url);}else {var Q$8=f9Cm$.Y87;Q$8+=f9Cm$.l60;Q$8+=Q6Q;ajaxOpts[Q$8]=url;}$[s5O]($[d7H](ajaxOpts,{data:data,success:update}));}});return this;}function destroy(){var X2k="un";var b$T="dte";var Y1K="emplate";var u9l="empl";var p15="estro";var K1X=X2k;K1X+=K18;K1X+=R2$;K1X+=X3P;var v6v=u5z;v6v+=b$T;var Y3d=f9Cm$.J4L;Y3d+=u9l;Y3d+=o2e;if(this[f9Cm$.J3V][n5R]){this[H9B]();}this[M1r]();if(this[f9Cm$.J3V][Y3d]){var T0e=f9Cm$.J4L;T0e+=Y1K;var o1P=f9Cm$.e08;o1P+=w1H;o1P+=f9Cm$[555616];var H2S=r0K;H2S+=d0b;$(H2S)[o1P](this[f9Cm$.J3V][T0e]);}var controller=this[f9Cm$.J3V][I2G];if(controller[J8j]){var f6b=f9Cm$[555616];f6b+=p15;f6b+=g5f;controller[f6b](this);}$(document)[v93](v6v + this[f9Cm$.J3V][K1X]);this[X6t]=B3c;this[f9Cm$.J3V]=B3c;}function disable(name){var Z9a="_fieldName";var t8z=Z9a;t8z+=f9Cm$.J3V;var that=this;$[a8N](this[t8z](name),function(i,n){var d4u="isable";I8Z.m_d();var l_b=f9Cm$[555616];l_b+=d4u;that[D_P](n)[l_b]();});return this;}function display(showIn){var a14="played";var q0J=M_R;q0J+=J_E;var R6G=f9Cm$[23424];R6G+=B1d;R6G+=f9Cm$.t_T;R6G+=f9Cm$[481343];I8Z.m_d();if(showIn === undefined){var X5l=h44;X5l+=f9Cm$.J3V;X5l+=a14;return this[f9Cm$.J3V][X5l];}return this[showIn?R6G:q0J]();}function displayed(){var A1k=f9Cm$[228782];A1k+=f$A;A1k+=X5O;var Z_5=D6p;Z_5+=f9Cm$.e08;Z_5+=B1d;I8Z.m_d();return $[Z_5](this[f9Cm$.J3V][A1k],function(fieldIn,name){I8Z.m_d();return fieldIn[n5R]()?name:B3c;});}function displayNode(){var C7Q="ler";var M7Y=f9Cm$[481343];M7Y+=f9Cm$[23424];M7Y+=f9Cm$[555616];I8Z.m_d();M7Y+=f9Cm$.t_T;var O9D=G7z;O9D+=s0$;O9D+=C7Q;return this[f9Cm$.J3V][O9D][M7Y](this);}function edit(items,arg1,arg2,arg3,arg4){var O_8="_crudA";var i2A="_data";var y75=D_P;y75+=f9Cm$.J3V;var E0N=i2A;E0N+=t0O;E0N+=f9Cm$[23424];E0N+=P7B;var E3b=O_8;E3b+=f9Cm$.l60;E3b+=M8F;var _this=this;var that=this;if(this[M1B](function(){I8Z.m_d();var e$b=w3$;e$b+=K18;e$b+=f9Cm$.J4L;that[e$b](items,arg1,arg2,arg3,arg4);})){return this;}var argOpts=this[E3b](arg1,arg2,arg3,arg4);this[p9D](items,this[E0N](y75,items),m0M,argOpts[w3u],function(){var F6Z="_assemble";var Y8W="Main";var J5w="maybeOpen";var n0q=J_q;n0q+=f9Cm$.E9M;var s6h=F6Z;s6h+=Y8W;_this[s6h]();_this[Q49](argOpts[n0q]);argOpts[J5w]();});return this;}function enable(name){var w0R=v7I;w0R+=Z4u;var that=this;$[a8N](this[w0R](name),function(i,n){that[D_P](n)[y9q]();});return this;}function error$1(name,msg){var q49="globalErr";var h7y=f9Cm$[555616];h7y+=f9Cm$[23424];h7y+=D6p;var wrapper=$(this[h7y][A2A]);if(msg === undefined){var E9c=q49;E9c+=A5h;var K9t=o9d;K9t+=D6p;var t7U=j7g;t7U+=c1_;this[t7U](this[K9t][c2Q],name,X17,function(){var Y5s="nF";var e3L="ggle";var x2U="ormError";var r3R=K18;r3R+=Y5s;r3R+=x2U;var W3l=H$0;W3l+=e3L;W3l+=E86;wrapper[W3l](r3R,name !== undefined && name !== j_l);});this[f9Cm$.J3V][E9c]=name;}else {var Z0k=y3N;Z0k+=A5h;this[D_P](name)[Z0k](msg);}return this;}function field(name){var p32="kn";var j7P="own field n";var l7P="Un";var n8z="ame - ";var e60=f9Cm$[228782];e60+=K18;e60+=f9Cm$.t_T;e60+=o$D;var sFields=this[f9Cm$.J3V][e60];if(!sFields[name]){var z8Z=l7P;z8Z+=p32;z8Z+=j7P;z8Z+=n8z;throw new Error(z8Z + name);}return sFields[name];}function fields(){var e7u=D6p;e7u+=C$g;return $[e7u](this[f9Cm$.J3V][P5P],function(fieldIn,name){return name;});}function file(name,id){var H9U='Unknown file id ';var Q3g=' in table ';var O9s=f9Cm$[228782];O9s+=K18;O9s+=G0R;var tableFromFile=this[O9s](name);I8Z.j$H();var fileFromTable=tableFromFile[id];if(!fileFromTable){throw new Error(H9U + id + Q3g + name);}return tableFromFile[id];}function files(name){var j$A='Unknown file table name: ';var r2P=f9Cm$[228782];r2P+=K18;r2P+=k_d;r2P+=f9Cm$.J3V;if(!name){return Editor[q2Y];}var editorTable=Editor[r2P][name];if(!editorTable){throw new Error(j$A + name);}return editorTable;}function get(name){I8Z.m_d();var q6F=h7H;q6F+=C6V;var O6P=s4L;O6P+=Q6Q;O6P+=f9Cm$[555616];var that=this;if(!name){var R09=F3o;R09+=K5X;name=this[R09]();}if(Array[d2_](name)){var out_1={};$[a8N](name,function(i,n){var I48=P9U;I48+=f9Cm$.J4L;var I9B=F3o;I8Z.j$H();I9B+=f9Cm$.t_T;I9B+=Q6Q;I9B+=f9Cm$[555616];out_1[n]=that[I9B](n)[I48]();});return out_1;}return this[O6P](name)[q6F]();}function hide(names,animate){var o4j=v7I;o4j+=Z4u;var J76=q4u;J76+=q9Z;J76+=s$2;var that=this;$[J76](this[o4j](names),function(i,n){var I9T=s$2;I9T+=U7A;I9T+=f9Cm$.t_T;var P0u=f9Cm$[228782];P0u+=K18;P0u+=f9Cm$.t_T;P0u+=e30;that[P0u](n)[I9T](animate);});return this;}function ids(includeHash){var u$2=G3n;u$2+=c_m;u$2+=s4W;u$2+=o$D;I8Z.m_d();if(includeHash === void C37){includeHash=h4R;}return $[X5_](this[f9Cm$.J3V][u$2],function(editIn,idSrc){return includeHash === X17?H10 + idSrc:idSrc;});}function inError(inNames){var T_F="rmE";var w4Y="globalError";var q$K=D3W;q$K+=h7H;q$K+=f9Cm$.J4L;I8Z.m_d();q$K+=s$2;var m80=f9Cm$[228782];m80+=f9Cm$[23424];m80+=T_F;m80+=o5Z;$(this[X6t][m80]);if(this[f9Cm$.J3V][w4Y]){return X17;}var names=this[Q_4](inNames);for(var i=C37,ien=names[q$K];i < ien;i++){if(this[D_P](names[i])[G9w]()){return X17;}}return h4R;}function inline(cell,fieldName,opts){var Q5i="idy";var W_0="DTE_Field";var d$F='Cannot edit more than one row inline at a time';var k5z="ainObject";var l6X="Options";var s1q="line";I8Z.m_d();var F4g="aSource";var i0D="ys";var L9R=B96;L9R+=s1q;var k1h=j7g;k1h+=w3$;k1h+=g8Y;var X0r=b7w;X0r+=Q5i;var W79=Q6Q;W79+=q8V;W79+=f9Cm$.J4L;W79+=s$2;var R5F=B2B;R5F+=W_0;var c4I=f4W;c4I+=s$2;var M3P=i3K;M3P+=i0D;var E9$=j7g;E9$+=o7y;E9$+=F4g;var c9t=g2v;c9t+=b1Z;c9t+=l6X;var y30=f9Cm$.t_T;y30+=G1H;y30+=R6y;y30+=f9Cm$[555616];var w9J=o27;w9J+=k5z;var _this=this;var that=this;if($[w9J](fieldName)){opts=fieldName;fieldName=undefined;}opts=$[y30]({},this[f9Cm$.J3V][c9t][n0T],opts);var editFields=this[E9$](k8T,cell,fieldName);var keys=Object[M3P](editFields);if(keys[c4I] > Y5Y){throw new Error(d$F);}var editRow=editFields[keys[C37]];var hosts=[];for(var _i=C37,_a=editRow[P8n];_i < _a[B97];_i++){var J64=d0i;J64+=s$2;var row=_a[_i];hosts[J64](row);}if($(R5F,hosts)[W79]){return this;}if(this[X0r](function(){var Y3z="inlin";var g_T=Y3z;g_T+=f9Cm$.t_T;that[g_T](cell,fieldName,opts);})){return this;}this[k1h](cell,editFields,L9R,opts,function(){I8Z.j$H();_this[Q$W](editFields,opts);});return this;}function inlineCreate(insertPoint,opts){var h5B="tionClass";var B2I="itF";var y33="Creat";var N70='fakeRow';var F9q=Y_$;F9q+=f9Cm$.J4L;F9q+=y33;F9q+=f9Cm$.t_T;var m_C=w3$;m_C+=B2I;m_C+=s4W;m_C+=o$D;var o0w=I1O;o0w+=h5B;var R5J=b2n;R5J+=h8a;var z4d=p9s;z4d+=B96;var C7_=q1F;C7_+=f9Cm$[555616];C7_+=f9Cm$.J3V;var U7B=g4c;U7B+=d0b;var _this=this;if($[F7v](insertPoint)){opts=insertPoint;insertPoint=B3c;}if(this[U7B](function(){I8Z.m_d();_this[W6L](insertPoint,opts);})){return this;}$[a8N](this[f9Cm$.J3V][C7_],function(name,fieldIn){var o5i="iReset";var B5F=f9Cm$.J3V;B5F+=C6V;var Z0D=D6p;Z0D+=f9Cm$.Y87;Z0D+=O6R;Z0D+=o5i;fieldIn[Z0D]();fieldIn[Y$8](C37,fieldIn[Z87]());fieldIn[B5F](fieldIn[Z87]());});this[f9Cm$.J3V][h56]=z4d;this[f9Cm$.J3V][R5J]=c2D;this[f9Cm$.J3V][n9R]=B3c;this[f9Cm$.J3V][t67]=this[e3V](N70,insertPoint);opts=$[d7H]({},this[f9Cm$.J3V][U4K][n0T],opts);this[o0w]();this[Q$W](this[f9Cm$.J3V][m_C],opts,function(){var u56="ource";var L1s="fa";I8Z.j$H();var K_9="_dataS";var l8K="keRowEnd";var T4N=L1s;T4N+=l8K;var m54=K_9;m54+=u56;_this[m54](T4N);});this[G29](F9q,B3c);return this;}function message(name,msg){var V$b="mInfo";if(msg === undefined){var s8C=g2v;s8C+=f9Cm$.l60;s8C+=V$b;var i$H=f9Cm$[555616];i$H+=f9Cm$[23424];i$H+=D6p;var U4y=t9a;U4y+=s8A;this[U4y](this[i$H][s8C],name);}else {var E4D=D6p;E4D+=B12;E4D+=f9Cm$.J3V;E4D+=D2Y;var U7J=s4L;U7J+=e30;this[U7J](name)[E4D](msg);}return this;}function mode(modeIn){var J0u='Not currently in an editing mode';var G0p='Changing from create mode is not supported';var g0R=b2n;g0R+=h8a;var y38=q9Z;y38+=f9Cm$.l60;y38+=f9Cm$.t_T;y38+=o2e;var P0A=f9Cm$.e08;P0A+=q9Z;P0A+=O74;if(!modeIn){return this[f9Cm$.J3V][Y8d];}if(!this[f9Cm$.J3V][Y8d]){throw new Error(J0u);}else if(this[f9Cm$.J3V][P0A] === y38 && modeIn !== c2D){throw new Error(G0p);}this[f9Cm$.J3V][g0R]=modeIn;I8Z.j$H();return this;}function modifier(){return this[f9Cm$.J3V][n9R];}function multiGet(fieldNames){var A$V="tiGet";var i32=M6F;i32+=Q6Q;i32+=A$V;var that=this;if(fieldNames === undefined){fieldNames=this[P5P]();}if(Array[d2_](fieldNames)){var out_2={};$[a8N](fieldNames,function(i,name){out_2[name]=that[D_P](name)[o1D]();});return out_2;}return this[D_P](fieldNames)[i32]();}function multiSet(fieldNames,valIn){var Q8h="sPlainOb";var G$_="ltiSet";var D_u=K18;D_u+=Q8h;D_u+=Q4H;var that=this;if($[D_u](fieldNames) && valIn === undefined){$[a8N](fieldNames,function(name,value){var z9R=C2U;I8Z.m_d();z9R+=C2h;z9R+=C6V;that[D_P](name)[z9R](value);});}else {var U2r=M6F;U2r+=G$_;var c9P=F3o;c9P+=G6S;c9P+=f9Cm$[555616];this[c9P](fieldNames)[U2r](valIn);}return this;}function node(name){var N$_=F3o;N$_+=D9y;I8Z.j$H();var a84=D6p;a84+=f9Cm$.e08;a84+=B1d;var that=this;if(!name){name=this[U2b]();}return Array[d2_](name)?$[a84](name,function(n){var I7K=L6$;I8Z.m_d();I7K+=f9Cm$.t_T;var r_o=f9Cm$[228782];r_o+=f$A;r_o+=f9Cm$[555616];return that[r_o](n)[I7K]();}):this[N$_](name)[z$j]();}function off(name,fn){$(this)[v93](this[A1x](name),fn);return this;}function on(name,fn){var G$2="tName";var c2b=p3g;c2b+=Q$x;c2b+=G$2;var O7q=f9Cm$[23424];O7q+=f9Cm$[481343];$(this)[O7q](this[c2b](name),fn);return this;}function one(name,fn){I8Z.m_d();var l_x=f9Cm$[23424];l_x+=f9Cm$[481343];l_x+=f9Cm$.t_T;$(this)[l_x](this[A1x](name),fn);return this;}function open(){var D_x="oseReg";var i9B="_c";var M4C="nest";var u_$=q_Z;u_$+=O4Y;u_$+=N4D;var q7U=i9B;q7U+=Q6Q;q7U+=D_x;var _this=this;this[Z4X]();this[q7U](function(){var W52="Close";var c2g="_nest";var W3j=c2g;W3j+=f9Cm$.t_T;W3j+=f9Cm$[555616];I8Z.j$H();W3j+=W52;_this[W3j](function(){var O3b="amicInfo";var y6h="earDyn";var c2x=i9B;c2x+=Q6Q;c2x+=y6h;c2x+=O3b;_this[c2x]();_this[G29](u7C,[m0M]);});});var ret=this[u_$](m0M);if(!ret){return this;}this[Y1l](function(){var u7$="rd";var F3w=d0M;F3w+=M42;var W0N=C6T;W0N+=q9H;W0N+=f9Cm$.J4L;W0N+=f9Cm$.J3V;var z5z=f9Cm$[23424];z5z+=u7$;z5z+=F7I;var j9W=v7I;j9W+=f9Cm$[23424];I8Z.j$H();j9W+=e4v;j9W+=f9Cm$.J3V;_this[j9W]($[X5_](_this[f9Cm$.J3V][z5z],function(name){var e$k=f9Cm$[228782];I8Z.j$H();e$k+=f$A;e$k+=f9Cm$[555616];e$k+=f9Cm$.J3V;return _this[f9Cm$.J3V][e$k][name];}),_this[f9Cm$.J3V][W0N][F3w]);_this[G29](I27,[m0M,_this[f9Cm$.J3V][Y8d]]);},this[f9Cm$.J3V][x6t][M4C]);this[s3c](m0M,h4R);return this;}function order(setIn){var H6Y="rder";var s7n='All fields, and no additional fields, must be provided for ordering.';var m2b="rt";var P1_="so";var p0Z="_displayRe";var O3R=p0Z;O3R+=U2b;var p$S=f9Cm$[23424];p$S+=H6Y;var F$h=X2q;F$h+=f9Cm$.J4L;F$h+=n3_;F$h+=f9Cm$[555616];var h5q=P1_;h5q+=f9Cm$.l60;h5q+=f9Cm$.J4L;var k$_=f9Cm$.J3V;k$_+=f9Cm$[23424];k$_+=m2b;var E6c=f9Cm$.J3V;E6c+=Q6Q;E6c+=B1B;E6c+=f9Cm$.t_T;var f38=A5h;f38+=R3S;I8Z.j$H();f38+=f9Cm$.l60;if(!setIn){var k4D=G5D;k4D+=F7I;return this[f9Cm$.J3V][k4D];}if(arguments[B97] && !Array[d2_](setIn)){setIn=Array[P99][k3q][B7t](arguments);}if(this[f9Cm$.J3V][f38][E6c]()[k$_]()[R5X](p69) !== setIn[k3q]()[h5q]()[R5X](p69)){throw new Error(s7n);}$[F$h](this[f9Cm$.J3V][p$S],setIn);this[O3R]();return this;}function remove(items,arg1,arg2,arg3,arg4){var L7I="difie";var S8F="_cr";var e6o='fields';var F4u="_act";var M33="udA";var P73="tRemove";var P4r=f9Cm$[481343];P4r+=x4Y;var i15=Y_$;i15+=P73;var x8H=F4u;x8H+=K18;x8H+=h8a;x8H+=E86;var C8H=f9Cm$.J3V;C8H+=f9Cm$.J4L;C8H+=g5f;C8H+=k_d;var F$l=g2v;F$l+=b1Z;var Q9V=D6p;Q9V+=f9Cm$[23424];Q9V+=L7I;Q9V+=f9Cm$.l60;var J7Z=u6E;J7Z+=K18;J7Z+=h8a;var O7t=S8F;O7t+=M33;O7t+=f9Cm$.l60;O7t+=M8F;var E17=f9Cm$.J4L;E17+=H3O;E17+=f9Cm$.t_T;var _this=this;var that=this;if(this[M1B](function(){var w7y=O4Y;w7y+=D6p;w7y+=l$I;w7y+=f9Cm$.t_T;I8Z.m_d();that[w7y](items,arg1,arg2,arg3,arg4);})){return this;}if(!items && !this[f9Cm$.J3V][E17]){var t2y=i3K;t2y+=z7F;t2y+=B12;t2y+=f9Cm$.J3V;items=t2y;}if(items[B97] === undefined){items=[items];}var argOpts=this[O7t](arg1,arg2,arg3,arg4);var editFields=this[e3V](e6o,items);this[f9Cm$.J3V][J7Z]=m8Z;this[f9Cm$.J3V][Q9V]=items;this[f9Cm$.J3V][t67]=editFields;this[X6t][F$l][C8H][G7z]=n50;this[x8H]();this[G29](i15,[pluck(editFields,P4r),pluck(editFields,v$r),items],function(){var U6Z="ltiRemove";var p6o="initMu";var W8c=p6o;W8c+=U6Z;I8Z.j$H();_this[G29](W8c,[editFields,items],function(){var L3D="ocus";var F4W="Ope";var f9t="pt";var Z8u="editO";var u2f="_assembleMain";var O2H="_fo";var Q1f="maybe";var H8f="rmO";var V$1=Z8u;V$1+=f9t;V$1+=f9Cm$.J3V;var l4D=Q1f;l4D+=F4W;l4D+=f9Cm$[481343];var y0k=f9Cm$[23424];y0k+=B1d;y0k+=f9Cm$.E9M;var J4w=O2H;J4w+=H8f;J4w+=C8a;_this[u2f]();_this[J4w](argOpts[y0k]);argOpts[l4D]();var opts=_this[f9Cm$.J3V][V$1];if(opts[G0n] !== B3c){var h3v=f9Cm$[228782];h3v+=L3D;var X9N=f9Cm$.t_T;X9N+=R2$;var x5_=q1i;x5_+=C_v;x5_+=f9c;var u3Y=o9d;u3Y+=D6p;var c2$=J0n;c2$+=f9Cm$.Y87;c2$+=f9Cm$.J4L;c2$+=L_M;$(c2$,_this[u3Y][x5_])[X9N](opts[h3v])[G0n]();}});});return this;}function set(setIn,valIn){var r_D="je";var Q6x="inOb";var W6_=M4P;W6_+=Q6x;W6_+=r_D;W6_+=Q2A;var that=this;if(!$[W6_](setIn)){var o={};o[setIn]=valIn;setIn=o;}$[a8N](setIn,function(n,v){I8Z.m_d();that[D_P](n)[D0u](v);});I8Z.m_d();return this;}function show(names,animate){var that=this;$[a8N](this[Q_4](names),function(i,n){var e7i="how";var K2T=f9Cm$.J3V;K2T+=e7i;that[D_P](n)[K2T](animate);});return this;}function submit(successCallback,errorCallback,formatdata,hideIn){var h30='div.DTE_Field';var v7T="activeEl";var w11=f9Cm$.t_T;w11+=f9Cm$.e08;w11+=j1L;var T7M=f9Cm$.t_T;T7M+=f9Cm$.e08;T7M+=q9Z;T7M+=s$2;var d23=F7I;d23+=a_T;d23+=f9Cm$.l60;var o6d=k_d;o6d+=L47;o6d+=f9Cm$.J4L;o6d+=s$2;var E_I=q9Z;E_I+=F7s;E_I+=P_W;E_I+=f9Cm$.J4L;var S5W=v7T;S5W+=f9Cm$.t_T;S5W+=Z22;var N4w=q2t;N4w+=f9Cm$[481343];var x8e=q1F;x8e+=X5O;var _this=this;var fields=this[f9Cm$.J3V][x8e];var errorFields=[];var errorReady=C37;var sent=h4R;if(this[f9Cm$.J3V][T0W] || !this[f9Cm$.J3V][N4w]){return this;}this[I0J](X17);var send=function(){var n26='initSubmit';var t3h=j7g;t3h+=o6f;t3h+=f9Cm$.t_T;t3h+=l1E;if(errorFields[B97] !== errorReady || sent){return;}_this[t3h](n26,[_this[f9Cm$.J3V][Y8d]],function(result){var Q8k=J04;Q8k+=n2z;Q8k+=S_t;if(result === h4R){_this[I0J](h4R);return;}I8Z.m_d();sent=X17;_this[Q8k](successCallback,errorCallback,formatdata,hideIn);});};var active=document[S5W];if($(active)[E_I](h30)[o6d] !== C37){active[h1_]();}this[d23]();$[T7M](fields,function(name,fieldIn){I8Z.m_d();if(fieldIn[G9w]()){errorFields[Z_J](name);}});$[w11](errorFields,function(i,name){var H7d=f9Cm$.t_T;I8Z.j$H();H7d+=f9Cm$.l60;H7d+=f9Cm$.l60;H7d+=A5h;fields[name][H7d](j_l,function(){I8Z.m_d();errorReady++;send();});});send();I8Z.m_d();return this;}function table(setIn){var o78=f9Cm$.J4L;o78+=f9Cm$.e08;o78+=J0n;o78+=k_d;if(setIn === undefined){return this[f9Cm$.J3V][z6u];}this[f9Cm$.J3V][o78]=setIn;return this;}function template(setIn){var n56="mpl";if(setIn === undefined){var L$o=U58;L$o+=n56;L$o+=f9Cm$.e08;L$o+=U58;return this[f9Cm$.J3V][L$o];}I8Z.m_d();this[f9Cm$.J3V][b2T]=setIn === B3c?B3c:$(setIn);return this;}function title(titleIn){var D8A="ddC";var j_I="ead";var P$2="lasse";var U8Y="class";var T8Z=B4Y;T8Z+=f9Cm$.J4L;T8Z+=k_d;var V3w=h0z;V3w+=Q6Q;var p5T=f9Cm$.e08;p5T+=D8A;p5T+=S81;var u2_=i$Z;u2_+=k8A;u2_+=R$r;var G2A=f9Cm$.J4L;G2A+=f9Cm$.e08;G2A+=h7H;var w5H=f9Cm$.J4L;w5H+=g8Y;w5H+=Q6Q;w5H+=f9Cm$.t_T;var V7Z=s$2;V7Z+=j_I;V7Z+=f9Cm$.t_T;V7Z+=f9Cm$.l60;var W_G=q9Z;W_G+=P$2;W_G+=f9Cm$.J3V;var header=$(this[X6t][L2d])[n_U](v1_ + this[W_G][L2d][P8l]);var titleClass=this[A2x][V7Z][w5H];if(titleIn === undefined){var T9d=f9Cm$.J4L;T9d+=K18;T9d+=f9Cm$.J4L;T9d+=k_d;var k2z=f9Cm$[555616];k2z+=f9Cm$.e08;k2z+=f9Cm$.J4L;k2z+=f9Cm$.e08;return header[k2z](T9d);}if(typeof titleIn === t0t){var G2_=Y4w;G2_+=f9Cm$.t_T;titleIn=titleIn(this,new DataTable$5[s5A](this[f9Cm$.J3V][G2_]));}var set=titleClass[G2A]?$(k8A + titleClass[R$f] + u2_ + titleClass[R$f])[p5T](titleClass[U8Y])[V3w](titleIn):titleIn;header[n8n](set)[I$4](T8Z,titleIn);return this;}function val(fieldIn,value){I8Z.m_d();var J$L="bj";var P5O="isPlainO";var o1Q=P9U;o1Q+=f9Cm$.J4L;var s1P=P5O;s1P+=J$L;s1P+=M$g;if(value !== undefined || $[s1P](fieldIn)){var R3H=f9Cm$.J3V;R3H+=C6V;return this[R3H](fieldIn,value);}return this[o1Q](fieldIn);;}function error(msg,tn,thro){var y$i="or more informatio";var b6N=" F";var M8u="n, please refer to https://";var c6I="warn";var p_7="datatables.net/tn/";var K9o=b6N;K9o+=y$i;K9o+=M8u;K9o+=p_7;if(thro === void C37){thro=X17;}var display=tn?msg + K9o + tn:msg;if(thro){throw display;}else {console[c6I](display);}}function pairs(data,props,fn){var V29=S1i;V29+=J0n;V29+=f9Cm$.t_T;V29+=Q6Q;var i;var ien;var dataPoint;props=$[d7H]({label:V29,value:T1q},props);if(Array[d2_](data)){for((i=C37,ien=data[B97]);i < ien;i++){dataPoint=data[i];if($[F7v](dataPoint)){var H6d=k6$;H6d+=f9Cm$.J4L;H6d+=f9Cm$.l60;fn(dataPoint[props[l8T]] === undefined?dataPoint[props[e9d]]:dataPoint[props[l8T]],dataPoint[props[e9d]],i,dataPoint[H6d]);}else {fn(dataPoint,dataPoint,i);}}}else {i=C37;$[a8N](data,function(key,val){fn(val,key,i);i++;});}}function upload$1(editor,conf,files,progressCallback,completeCallback){var e_G="errors";var i40='<i>Uploading file</i>';var m5u="spli";var H$P="_limitLeft";var W2v='A server error occurred while uploading the file';var v_G="aUR";var s4P="onloa";var X_5="fileReadText";var w4M="nctio";var B$H="readAsDa";var a9Q=B$H;a9Q+=f9Cm$.J4L;a9Q+=v_G;a9Q+=n35;var n$Q=D6p;n$Q+=f9Cm$.e08;n$Q+=B1d;var o8F=s4P;o8F+=f9Cm$[555616];var e63=f9Cm$[228782];e63+=f9Cm$.Y87;e63+=w4M;e63+=f9Cm$[481343];var z1s=P9q;z1s+=f9Cm$.e08;z1s+=G1H;var K5r=f9Cm$[481343];K5r+=o1Y;var f2o=F7I;f2o+=f9Cm$.l60;f2o+=f9Cm$[23424];f2o+=f9Cm$.l60;var G8r=f9Cm$.t_T;G8r+=f9Cm$.l60;G8r+=h5o;G8r+=f9Cm$.J3V;var reader=new FileReader();var counter=C37;var ids=[];var generalError=conf[G8r] && conf[e_G][j7g]?conf[e_G][j7g]:W2v;editor[f2o](conf[K5r],j_l);if(typeof conf[z1s] === e63){var r1a=f9Cm$.e08;r1a+=Q23;r1a+=G1H;conf[r1a](files,function(idsIn){I8Z.j$H();completeCallback[B7t](editor,idsIn);});return;}progressCallback(conf,conf[X_5] || i40);reader[o8F]=function(e){var m9v="load plug-in";var d8W="pload";var o1w='preUpload';var n$X="No Ajax option specified";var T9F="ajaxData";var d6r='uploadField';var l0B="Data";var p8C="aja";var z4J="`ajax.data` with an object. Please use it as a function instead.";var B_O=" for ";var E0R='upload';var I2n="ad feature cannot use ";var F3E="Uplo";var E2E=j7g;E2E+=f9Cm$.t_T;E2E+=a6R;E2E+=l1E;var I5C=f9Cm$[555616];I5C+=f9Cm$.e08;I5C+=f9Cm$.J4L;I5C+=f9Cm$.e08;var V4h=f9Cm$[326480];V4h+=f9Cm$.J4L;V4h+=f9Cm$.e08;var p1U=P9q;p1U+=J_Y;var e8e=f9Cm$.e08;e8e+=w1H;e8e+=f9Cm$[555616];var L2L=f9Cm$.e08;L2L+=Y8i;L2L+=n3_;L2L+=f9Cm$[555616];var O9v=f9Cm$.Y87;O9v+=d8W;var P7Y=u6E;P7Y+=X3e;var data=new FormData();var ajax;data[P2$](P7Y,O9v);data[L2L](d6r,conf[h2d]);data[e8e](E0R,files[counter]);if(conf[T9F]){var l5T=f9Cm$.e08;l5T+=f9Cm$.Z$r;l5T+=J_Y;l5T+=l0B;conf[l5T](data,files[counter],counter);}if(conf[p1U]){var n_T=p8C;n_T+=G1H;ajax=conf[n_T];}else if($[F7v](editor[f9Cm$.J3V][N7l])){var Q$I=f9Cm$.Y87;Q$I+=r3r;Q$I+=O1b;var q60=f9Cm$.e08;q60+=f9Cm$.Z$r;q60+=f9Cm$.e08;q60+=G1H;var D1a=f9Cm$.Y87;D1a+=d8W;var Y9C=P9q;Y9C+=f9Cm$.e08;Y9C+=G1H;ajax=editor[f9Cm$.J3V][Y9C][D1a]?editor[f9Cm$.J3V][q60][Q$I]:editor[f9Cm$.J3V][N7l];}else if(typeof editor[f9Cm$.J3V][N7l] === t3X){ajax=editor[f9Cm$.J3V][N7l];}if(!ajax){var t$9=n$X;t$9+=B_O;t$9+=j7q;t$9+=m9v;throw new Error(t$9);}if(typeof ajax === t3X){ajax={url:ajax};}if(typeof ajax[V4h] === t0t){var L2f=f9Cm$.t_T;L2f+=L7Y;var i3v=f9Cm$.J3V;i3v+=f9Cm$.J4L;i3v+=o46;i3v+=L47;var d={};var ret=ajax[I$4](d);if(ret !== undefined && typeof ret !== i3v){d=ret;}$[L2f](d,function(key,value){I8Z.m_d();data[P2$](key,value);});}else if($[F7v](ajax[I5C])){var z1k=F3E;z1k+=I2n;z1k+=z4J;throw new Error(z1k);}editor[E2E](o1w,[conf[h2d],files[counter],data],function(preRet){var c2X="upl";var G0y="readAsDataURL";var v__='preSubmit.DTE_Upload';var w$$='post';var z3v=X2q;z3v+=N3C;var M9L=f9Cm$.e08;I8Z.j$H();M9L+=f9Cm$.Z$r;M9L+=J_Y;var d_A=f9Cm$[23424];d_A+=f9Cm$[481343];if(preRet === h4R){var l5j=Q6Q;l5j+=f9Cm$.t_T;l5j+=f9Cm$[481343];l5j+=h5h;if(counter < files[l5j] - Y5Y){counter++;reader[G0y](files[counter]);}else {completeCallback[B7t](editor,ids);}return;}var submit=h4R;editor[d_A](v__,function(){I8Z.m_d();submit=X17;return h4R;});$[M9L]($[z3v]({},ajax,{contentType:h4R,data:data,dataType:h48,error:function(xhr){var m$R="bmit.DTE_Upload";var F58="stat";var H3z='uploadXhrError';var U1E="atu";var o$W=U5a;o$W+=U1E;o$W+=f9Cm$.J3V;var Z9N=F58;Z9N+=M42;var f6Y=f9Cm$.t_T;f6Y+=f9Cm$.l60;f6Y+=a_T;f6Y+=f9Cm$.l60;I8Z.j$H();var f4m=B1d;f4m+=K$K;f4m+=m$R;var s0w=f9Cm$[23424];s0w+=f9Cm$[228782];s0w+=f9Cm$[228782];var errors=conf[e_G];editor[s0w](f4m);editor[f6Y](conf[h2d],errors && errors[xhr[Z9N]]?errors[xhr[o$W]]:generalError);editor[G29](H3z,[conf[h2d],xhr]);progressCallback(conf);},processData:h4R,success:function(json){var x2r="ploa";I8Z.j$H();var S_3="preSubmit.DTE_U";var M3A="oadXhrSuccess";var A99="ldErrors";var c98=K18;c98+=f9Cm$[555616];var D1s=f9Cm$.Y87;D1s+=d8W;var x6Z=f9Cm$.Y87;x6Z+=B1d;x6Z+=v8S;var M10=Q6Q;M10+=n3_;M10+=h7H;M10+=K5v;var t1b=s4L;t1b+=A99;var a2A=j2Z;a2A+=f9Cm$[23424];a2A+=h_u;var b1E=f9Cm$[481343];b1E+=o1Y;var a32=c2X;a32+=M3A;var Y4u=p3g;Y4u+=a6R;Y4u+=f9Cm$[481343];Y4u+=f9Cm$.J4L;var G$I=S_3;G$I+=x2r;G$I+=f9Cm$[555616];var k5V=f9Cm$[23424];k5V+=o3v;editor[k5V](G$I);editor[Y4u](a32,[conf[b1E],json]);if(json[a2A] && json[t1b][M10]){var M7B=Q6Q;M7B+=f9Cm$.t_T;M7B+=y8m;var a0s=F81;a0s+=f9Cm$.J3V;var errors=json[a0s];for(var i=C37,ien=errors[M7B];i < ien;i++){var w2E=f9Cm$.t_T;w2E+=C58;w2E+=A5h;editor[w2E](errors[i][h2d],errors[i][C8k]);}completeCallback[B7t](editor,ids,X17);}else if(json[h8G]){var n7X=f9Cm$.t_T;n7X+=f9Cm$.l60;n7X+=h5o;var n9d=F7I;n9d+=h5o;editor[n9d](json[n7X]);completeCallback[B7t](editor,ids,X17);}else if(!json[x6Z] || !json[D1s][c98]){var L8N=f9Cm$.t_T;L8N+=C58;L8N+=f9Cm$[23424];L8N+=f9Cm$.l60;editor[L8N](conf[h2d],generalError);completeCallback[B7t](editor,ids,X17);}else {var j$m=K18;j$m+=f9Cm$[555616];var o9n=c2X;o9n+=f9Cm$[23424];o9n+=O1b;var Q5d=B1d;Q5d+=f9Cm$.Y87;Q5d+=f9Cm$.J3V;Q5d+=s$2;var K8_=f9Cm$[228782];K8_+=K18;K8_+=k_d;K8_+=f9Cm$.J3V;if(json[K8_]){var E8w=F3o;E8w+=G0R;$[a8N](json[E8w],function(table,filesIn){var Z8z=F3o;Z8z+=Q6Q;Z8z+=f9Cm$.t_T;Z8z+=f9Cm$.J3V;if(!Editor[q2Y][table]){var X2Z=f9Cm$[228782];X2Z+=K18;X2Z+=k_d;X2Z+=f9Cm$.J3V;Editor[X2Z][table]={};}$[d7H](Editor[Z8z][table],filesIn);});}ids[Q5d](json[o9n][j$m]);if(counter < files[B97] - Y5Y){counter++;reader[G0y](files[counter]);}else {var D5s=q9Z;D5s+=f9Cm$.e08;D5s+=Q6Q;D5s+=Q6Q;completeCallback[D5s](editor,ids);if(submit){var l2J=k18;l2J+=D6p;l2J+=K18;l2J+=f9Cm$.J4L;editor[l2J]();}}}progressCallback(conf);},type:w$$,xhr:function(){var E1B="oad";var H_U="ajaxSet";var v4W="onloade";var f1I="xhr";var N6n="nprogress";var R3l=j7q;R3l+=Q6Q;R3l+=E1B;var q_X=H_U;q_X+=z7B;var xhr=$[q_X][f1I]();if(xhr[R3l]){var O2S=v4W;O2S+=T52;var a37=f9Cm$.Y87;a37+=B1d;a37+=Q6Q;a37+=E1B;var n$P=f9Cm$[23424];n$P+=N6n;var K3x=c2X;K3x+=E1B;xhr[K3x][n$P]=function(e){var w4I='%';I8Z.j$H();var O3P="lengthComputable";var L39="total";var R1F="oa";var W9I="oFixed";var Y_f=100;var c8W=':';if(e[O3P]){var O6o=f9Cm$.J4L;O6o+=W9I;var m4D=Q6Q;m4D+=R1F;m4D+=f9Cm$[555616];m4D+=w3$;var percent=(e[m4D] / e[L39] * Y_f)[O6o](C37) + w4I;progressCallback(conf,files[B97] === Y5Y?percent:counter + c8W + files[B97] + E$d + percent);}};xhr[a37][O2S]=function(){var o9r="ssingTe";var Y6_="proce";var r7T="roc";var G6o=w_7;G6o+=r7T;I8Z.j$H();G6o+=z_k;G6o+=s29;var G0x=Y6_;G0x+=o9r;G0x+=M27;progressCallback(conf,conf[G0x] || G6o);};}return xhr;}}));});};files=$[n$Q](files,function(val){I8Z.j$H();return val;});if(conf[H$P] !== undefined){var D9q=Q6Q;D9q+=f9Cm$.t_T;D9q+=y8m;var O5r=m5u;O5r+=z8x;files[O5r](conf[H$P],files[D9q]);}reader[a9Q](files[C37]);}function factory(root,jq){var U$i="cum";var n_o="jqu";var K$d=n_o;K$d+=F7I;K$d+=g5f;var Z1V=o9d;Z1V+=U$i;Z1V+=u7S;I8Z.j$H();var is=h4R;if(root && root[Z1V]){window=root;document=root[f9Cm$.F4e];}if(jq && jq[f9Cm$.E4X] && jq[f9Cm$.E4X][K$d]){$=jq;is=X17;}return is;}var DataTable$4=$[f9Cm$.E4X][F40];var _inlineCounter=C37;function _actionClass(){var m$e="actions";var F63="jo";var q55="ddCl";var x5N="cre";var D27=f9Cm$.l60;D27+=f9Cm$.t_T;D27+=D6p;D27+=o3Z;var j_D=f9Cm$.t_T;j_D+=f9Cm$[555616];j_D+=K18;j_D+=f9Cm$.J4L;var U8u=x5N;U8u+=f9Cm$.e08;U8u+=f9Cm$.J4L;U8u+=f9Cm$.t_T;var s10=F63;s10+=K18;s10+=f9Cm$[481343];var t1Y=f9Cm$.t_T;t1Y+=f9Cm$[555616];t1Y+=g8Y;var Y2g=H3b;Y2g+=m0O;var m2t=f9Cm$.e08;m2t+=q9Z;m2t+=f9Cm$.J4L;m2t+=X3e;var classesActions=this[A2x][m$e];var action=this[f9Cm$.J3V][m2t];I8Z.m_d();var wrapper=$(this[X6t][Y2g]);wrapper[B_c]([classesActions[G75],classesActions[t1Y],classesActions[h6J]][s10](E$d));if(action === U8u){var i6T=q4o;i6T+=Q_A;i6T+=f9Cm$.t_T;var i60=f9Cm$.e08;i60+=q55;i60+=S$6;wrapper[i60](classesActions[i6T]);}else if(action === j_D){var R$1=f9Cm$.t_T;R$1+=h44;R$1+=f9Cm$.J4L;var w16=X0d;w16+=Q6Q;w16+=f9Cm$.e08;w16+=i3q;wrapper[w16](classesActions[R$1]);}else if(action === D27){wrapper[k1$](classesActions[h6J]);}}function _ajax(data,success,error,submitParams){var d7W="xtend";var X_p="complete";var m5v="ET";var y9f="split";var s5x="ram";var T$i="mple";var L1B="url";var H4N='POST';var h7_='?';var C6F=/{id}/;var D0z="replacements";var q9d="ift";var j6F=/_id_/;var r6V="uns";var C2x="omple";var R4S="place";var W5V="EL";var X0i="rl";var Q_H="ments";var S6$="deleteBody";var O87=f9Cm$.e08;O87+=f9Cm$.Z$r;O87+=f9Cm$.e08;O87+=G1H;var q7H=l93;q7H+=W5V;q7H+=m5v;q7H+=i2S;var K_a=f9Cm$.J4L;K_a+=g5f;K_a+=B1d;K_a+=f9Cm$.t_T;var f6l=f9Cm$[555616];f6l+=f9Cm$.e08;f6l+=f9Cm$.J4L;f6l+=f9Cm$.e08;var w_V=K2g;w_V+=H5i;var y2w=f9Cm$.Y87;y2w+=f9Cm$.l60;y2w+=Q6Q;var u9u=f9Cm$.Y87;u9u+=X0i;var o$H=f9Cm$.Z$r;o$H+=f9Cm$[23424];o$H+=K18;o$H+=f9Cm$[481343];var r9h=U7A;r9h+=t0O;r9h+=f9Cm$.l60;r9h+=q9Z;var q1N=C6T;q1N+=y7j;q1N+=g2a;q1N+=f9Cm$.J3V;var D$e=O4Y;D$e+=U7W;D$e+=f9Cm$.t_T;var G$4=f9Cm$.t_T;G$4+=f9Cm$[555616];G$4+=K18;G$4+=f9Cm$.J4L;var d73=f9Cm$.Z$r;d73+=f9Cm$.J3V;d73+=f9Cm$[23424];d73+=f9Cm$[481343];var g2Y=f9Cm$.e08;g2Y+=Q2A;g2Y+=X3e;var action=this[f9Cm$.J3V][g2Y];var thrown;var opts={complete:[function(xhr,text){var z4F="SON";var a1$=204;var T5H="responseJSON";var U5c="eText";var f5W="respons";var O0d="isPlainOb";var E5t="responseText";var p_k="responseJ";var J6L="parse";var y6y=400;var x2a=O0d;x2a+=Q4H;var v1l=f9Cm$[481343];v1l+=f9Cm$.Y87;v1l+=Q6Q;v1l+=Q6Q;var p7F=f5W;p7F+=U5c;var json=B3c;if(xhr[C8k] === a1$ || xhr[p7F] === v1l){json={};}else {try{var i5E=p_k;i5E+=z4F;json=xhr[T5H]?xhr[i5E]:JSON[J6L](xhr[E5t]);}catch(e){}}if($[x2a](json) || Array[d2_](json)){var Q2w=U5a;Q2w+=f9Cm$.e08;Q2w+=f9Cm$.J4L;Q2w+=M42;success(json,xhr[Q2w] >= y6y,xhr);}else {error(xhr,text,thrown);}}],data:B3c,dataType:d73,error:[function(xhr,text,err){I8Z.m_d();thrown=err;}],success:[],type:H4N};var a;var ajaxSrc=this[f9Cm$.J3V][N7l];var id=action === G$4 || action === D$e?pluck(this[f9Cm$.J3V][q1N],r9h)[o$H](U74):B3c;if($[F7v](ajaxSrc) && ajaxSrc[action]){ajaxSrc=ajaxSrc[action];}if(typeof ajaxSrc === t0t){var e1h=T93;e1h+=Q6Q;ajaxSrc[e1h](this,B3c,B3c,data,success,error);return;}else if(typeof ajaxSrc === t3X){if(ajaxSrc[x7i](E$d) !== -Y5Y){a=ajaxSrc[y9f](E$d);opts[c11]=a[C37];opts[L1B]=a[Y5Y];}else {var I99=f9Cm$.Y87;I99+=f9Cm$.l60;I99+=Q6Q;opts[I99]=ajaxSrc;}}else {var v3w=X2q;v3w+=R6y;v3w+=f9Cm$[555616];var m0N=f9Cm$.t_T;m0N+=C58;m0N+=f9Cm$[23424];m0N+=f9Cm$.l60;var Z$_=f9Cm$.t_T;Z$_+=d7W;var optsCopy=$[Z$_]({},ajaxSrc || ({}));if(optsCopy[X_p]){var U9U=q9Z;U9U+=C2x;U9U+=U58;var t9o=q9Z;t9o+=f9Cm$[23424];t9o+=T$i;t9o+=U58;opts[t9o][d9t](optsCopy[X_p]);delete optsCopy[U9U];}if(optsCopy[m0N]){var L1Y=f9Cm$.t_T;L1Y+=f9Cm$.l60;L1Y+=f9Cm$.l60;L1Y+=A5h;var B$r=f9Cm$.t_T;B$r+=C58;B$r+=f9Cm$[23424];B$r+=f9Cm$.l60;var t3y=r6V;t3y+=s$2;t3y+=q9d;var x66=t3Z;x66+=f9Cm$.l60;opts[x66][t3y](optsCopy[B$r]);delete optsCopy[L1Y];}opts=$[v3w]({},opts,optsCopy);}if(opts[D0z]){var l45=O4Y;l45+=R4S;l45+=Q_H;$[a8N](opts[l45],function(key,repl){var J$V="replac";var i7p='{';var P7g='}';var l3q=J$V;l3q+=f9Cm$.t_T;opts[L1B]=opts[L1B][l3q](i7p + key + P7g,repl[B7t](this,key,id,action,data));});}opts[u9u]=opts[y2w][w_V](j6F,id)[d9I](C6F,id);if(opts[I$4]){var h7E=f9Cm$.t_T;h7E+=M27;h7E+=f9Cm$.t_T;h7E+=T52;var e4m=f9Cm$[326480];e4m+=Q3e;var C1$=f9Cm$[228782];C1$+=f9Cm$.Y87;C1$+=f9Cm$[481343];C1$+=c7u;var isFn=typeof opts[I$4] === C1$;var newData=isFn?opts[I$4](data):opts[e4m];data=isFn && newData?newData:$[h7E](X17,data,newData);}opts[f6l]=data;if(opts[K_a] === q7H && (opts[S6$] === undefined || opts[S6$] === X17)){var s0d=f9Cm$[326480];s0d+=f9Cm$.J4L;s0d+=f9Cm$.e08;var t2F=f9Cm$.Y87;t2F+=X0i;var z0o=f9Cm$[555616];z0o+=q94;var C$Z=D6_;C$Z+=s5x;var params=$[C$Z](opts[z0o]);opts[t2F]+=opts[L1B][x7i](h7_) === -Y5Y?h7_ + params:s2o + params;delete opts[s0d];}$[O87](opts);}function _animate(target,style,time,callback){var v2o="anima";var O3f="cs";var y1x="nima";var O7g=f9Cm$.e08;O7g+=y1x;O7g+=f9Cm$.J4L;O7g+=f9Cm$.t_T;I8Z.j$H();if($[f9Cm$.E4X][O7g]){var j16=v2o;j16+=U58;target[Y4v]()[j16](style,time,callback);}else {var w76=O3f;w76+=f9Cm$.J3V;target[w76](style);var scope=target[B97] && target[B97] > Y5Y?target[C37]:target;if(typeof time === t0t){var h5F=q9Z;h5F+=f9Cm$.e08;h5F+=Q6Q;h5F+=Q6Q;time[h5F](scope);}else if(callback){callback[B7t](scope);}}}function _assembleMain(){var N3v="head";var T8O="bodyContent";var D$L="foo";var l2w="mErr";var W8m="rmInfo";var Y1R=i$c;I8Z.j$H();Y1R+=D6p;var O7T=f9Cm$.e08;O7T+=Q8$;O7T+=f9Cm$[481343];O7T+=f9Cm$[555616];var k3F=g2v;k3F+=W8m;var m8F=J0n;m8F+=U2j;m8F+=W4z;var P_C=f9Cm$[228782];P_C+=A5h;P_C+=l2w;P_C+=A5h;var V$x=p26;V$x+=b9X;var R2_=D$L;R2_+=A7I;var d$$=N3v;d$$+=f9Cm$.t_T;d$$+=f9Cm$.l60;var P6d=o9d;P6d+=D6p;var dom=this[P6d];$(dom[A2A])[g9P](dom[d$$]);$(dom[R2_])[V$x](dom[P_C])[P2$](dom[m8F]);$(dom[T8O])[P2$](dom[k3F])[O7T](dom[Y1R]);}function _blur(){var C6Q="onBlur";var J3L="Bl";var u5L="bm";var p7N=q9Z;p7N+=Q6Q;p7N+=f9Cm$[23424];p7N+=Y1m;var J2l=N5j;J2l+=u5L;J2l+=K18;J2l+=f9Cm$.J4L;var z1i=f9Cm$[235655];z1i+=f9Cm$[481343];z1i+=Q2A;z1i+=X3e;var K8C=I6h;K8C+=J3L;K8C+=f9Cm$.Y87;K8C+=f9Cm$.l60;var K6_=C6T;K6_+=q9H;K6_+=f9Cm$.J4L;I8Z.m_d();K6_+=f9Cm$.J3V;var opts=this[f9Cm$.J3V][K6_];var onBlur=opts[C6Q];if(this[G29](K8C) === h4R){return;}if(typeof onBlur === z1i){onBlur(this);}else if(onBlur === J2l){var y4l=k1a;y4l+=f9Cm$.J4L;this[y4l]();}else if(onBlur === p7N){this[R4h]();}}function _clearDynamicInfo(errorsOnly){var N04="iv.";var k9R="lasses";var U93=t3Z;U93+=f9Cm$.l60;var O5i=M61;O5i+=o6H;var Q0N=f9Cm$[555616];Q0N+=N04;var C5C=s4L;C5C+=Q6Q;C5C+=f9Cm$[555616];var H85=q9Z;H85+=k9R;if(errorsOnly === void C37){errorsOnly=h4R;}if(!this[f9Cm$.J3V]){return;}var errorClass=this[H85][C5C][h8G];var fields=this[f9Cm$.J3V][P5P];$(Q0N + errorClass,this[X6t][O5i])[B_c](errorClass);$[a8N](fields,function(name,field){var m2L=f9Cm$.t_T;m2L+=o5Z;field[m2L](j_l);if(!errorsOnly){var c0f=D6p;c0f+=s8A;field[c0f](j_l);}});this[U93](j_l);if(!errorsOnly){this[c1_](j_l);}}function _close(submitComplete,mode){var p3P="r-";var Y5_="closeCb";var h2w="seCb";var L1b="Cb";var r85="focus.edito";var d5x="ayed";var U42="oseIcb";var n8x="Ic";var N4_="osed";var J_k=M4_;J_k+=f9Cm$.J3V;J_k+=f9Cm$.t_T;var F5q=p3g;F5q+=a6R;F5q+=l1E;var n$e=s8o;n$e+=d5x;var w9B=r85;w9B+=p3P;w9B+=d0M;w9B+=M42;var c1P=f9Cm$[23424];c1P+=f9Cm$[228782];c1P+=f9Cm$[228782];var t7P=J0n;t7P+=q9c;t7P+=g5f;var C5B=q9Z;C5B+=h_t;C5B+=n8x;C5B+=J0n;var s5i=q9Z;s5i+=F7s;s5i+=h2w;var v4N=I6h;v4N+=s_b;v4N+=F7s;v4N+=Y1m;var N8S=y$s;N8S+=u7S;var closed;if(this[N8S](v4N) === h4R){return;}if(this[f9Cm$.J3V][s5i]){var P4U=q9Z;P4U+=F7s;P4U+=Y1m;P4U+=L1b;closed=this[f9Cm$.J3V][P4U](submitComplete,mode);this[f9Cm$.J3V][Y5_]=B3c;}if(this[f9Cm$.J3V][C5B]){var G0K=M_R;G0K+=U42;this[f9Cm$.J3V][c78]();this[f9Cm$.J3V][G0K]=B3c;}$(t7P)[c1P](w9B);this[f9Cm$.J3V][n$e]=h4R;this[F5q](J_k);if(closed){var P4K=M_R;P4K+=N4_;this[G29](P4K,[closed]);}}function _closeReg(fn){var p8J=q9Z;I8Z.m_d();p8J+=h_t;p8J+=s_b;p8J+=J0n;this[f9Cm$.J3V][p8J]=fn;}function _crudArgs(arg1,arg2,arg3,arg4){var V0w="sPlain";var e6b="main";var n1K=O94;n1K+=q9H;n1K+=Y$F;var x_J=f9Cm$.t_T;x_J+=G1H;x_J+=R6y;x_J+=f9Cm$[555616];var l6Y=K18;l6Y+=V0w;l6Y+=F6K;l6Y+=l$O;var that=this;var title;var buttons;var show;var opts;if($[l6Y](arg1)){opts=arg1;}else if(typeof arg1 === I06){show=arg1;opts=arg2;;}else {title=arg1;buttons=arg2;show=arg3;opts=arg4;;}if(show === undefined){show=X17;}if(title){var d37=B4Y;d37+=f9Cm$.J4L;d37+=Q6Q;d37+=f9Cm$.t_T;that[d37](title);}if(buttons){var q1o=J0n;q1o+=D07;q1o+=f9Cm$.J3V;that[q1o](buttons);}return {maybeOpen:function(){I8Z.j$H();if(show){that[N4D]();}},opts:$[x_J]({},this[f9Cm$.J3V][n1K][e6b],opts)};}function _dataSource(name){var e_7="dataSources";var g7A=f4W;g7A+=s$2;var args=[];for(var _i=Y5Y;_i < arguments[g7A];_i++){args[_i - Y5Y]=arguments[_i];}var dataSource=this[f9Cm$.J3V][z6u]?Editor[e_7][f9Cm$.m96]:Editor[e_7][n8n];var fn=dataSource[name];if(fn){return fn[D08](this,args);}}function _displayReorder(includeFields){var o1n="inclu";var T4y="rmC";var n10="yOrde";var C_3="deF";var f0a="dTo";var P4p=b2n;P4p+=h8a;var i4G=h44;i4G+=C3S;i4G+=n10;i4G+=f9Cm$.l60;var O$$=y$s;O$$+=f9Cm$.t_T;O$$+=l1E;var J7y=D6p;J7y+=f9Cm$.e08;J7y+=B96;var L4n=R3S;L4n+=f9Cm$.J4L;L4n+=f9Cm$.e08;L4n+=j1L;var q9t=F3o;q9t+=G6S;q9t+=X5O;var Y8X=g2v;Y8X+=T4y;Y8X+=Y1k;Y8X+=f9Cm$.J4L;var c5I=f9Cm$[555616];c5I+=f9Cm$[23424];c5I+=D6p;var _this=this;var formContent=$(this[c5I][Y8X]);var fields=this[f9Cm$.J3V][q9t];var order=this[f9Cm$.J3V][U2b];var template=this[f9Cm$.J3V][b2T];var mode=this[f9Cm$.J3V][h56] || m0M;if(includeFields){this[f9Cm$.J3V][u15]=includeFields;}else {var s$w=o1n;s$w+=C_3;s$w+=e6e;includeFields=this[f9Cm$.J3V][s$w];}formContent[n_U]()[L4n]();$[a8N](order,function(i,name){var d_U="eld[name=\"";var t3a="editor-fi";var v05="after";var V3G='[data-editor-template="';if(_this[f5n](name,includeFields) !== -Y5Y){if(template && mode === m0M){var z8C=f9Cm$.e08;z8C+=Y8i;z8C+=b9X;var a9N=f9Cm$[481343];a9N+=f9Cm$[23424];a9N+=f9Cm$[555616];a9N+=f9Cm$.t_T;var Z9t=y9e;Z9t+=M3E;var v3D=t3a;v3D+=d_U;var D6M=f9Cm$[228782];D6M+=K18;D6M+=f9Cm$[481343];D6M+=f9Cm$[555616];template[D6M](v3D + name + Z9t)[v05](fields[name][a9N]());template[I$n](V3G + name + D9g)[z8C](fields[name][z$j]());}else {formContent[P2$](fields[name][z$j]());}}});if(template && mode === J7y){var x6P=p26;x6P+=n3_;x6P+=f0a;template[x6P](formContent);}this[O$$](i4G,[this[f9Cm$.J3V][n5R],this[f9Cm$.J3V][P4p],formContent]);}function _edit(items,editFields,type,formOptions,setupDone){var L4J="ice";var N3e="lice";var V8E="tData";var d3A="itE";var M_z="ionClass";var x9L="sl";var z8q="nArra";var F9S="yReorder";var a1m=L6$;a1m+=f9Cm$.t_T;var p$6=B96;p$6+=d3A;p$6+=f9Cm$[555616];p$6+=g8Y;var G4T=j7g;G4T+=n_A;var N$N=j7g;N$N+=b7s;N$N+=S1i;N$N+=F9S;var o0_=x9L;o0_+=L4J;var T6i=I1O;T6i+=f9Cm$.J4L;T6i+=M_z;var c5H=D6p;c5H+=q9c;c5H+=f9Cm$.t_T;var t$3=R6r;t$3+=f9Cm$[23424];t$3+=K2u;var Z4f=G3n;Z4f+=V8E;var o89=f9Cm$[228782];o89+=g2a;I8Z.j$H();o89+=f9Cm$.J3V;var _this=this;var fields=this[f9Cm$.J3V][o89];var usedFields=[];var includeInOrder;var editData={};this[f9Cm$.J3V][t67]=editFields;this[f9Cm$.J3V][Z4f]=editData;this[f9Cm$.J3V][n9R]=items;this[f9Cm$.J3V][Y8d]=Z75;this[X6t][O94][r_x][G7z]=t$3;this[f9Cm$.J3V][c5H]=type;this[T6i]();$[a8N](fields,function(name,field){var S1u=Q6Q;S1u+=q8V;S1u+=K5v;var u3H=f9Cm$.t_T;I8Z.m_d();u3H+=o$n;u3H+=s$2;field[e3n]();includeInOrder=h4R;editData[name]={};$[u3H](editFields,function(idSrc,edit){var b_b="displayField";var S0k="displayFie";var G8d="faul";var T7D="layFields";var d1p="multiS";var t7w="nullDe";var g91=F3o;g91+=G6S;g91+=X5O;if(edit[g91][name]){var E1E=f9Cm$.l60;E1E+=f9Cm$[23424];E1E+=H3b;var V7Y=f9Cm$.J3V;V7Y+=q9Z;V7Y+=J_q;V7Y+=f9Cm$.t_T;var x4J=t7w;x4J+=G8d;x4J+=f9Cm$.J4L;var val=field[C$F](edit[I$4]);var nullDefault=field[x4J]();editData[name][idSrc]=val === B3c?j_l:Array[d2_](val)?val[k3q]():val;if(!formOptions || formOptions[V7Y] === E1E){var b0u=S0k;b0u+=o$D;var x2E=b_b;x2E+=f9Cm$.J3V;var g6B=d1p;g6B+=f9Cm$.t_T;g6B+=f9Cm$.J4L;field[g6B](idSrc,val === undefined || nullDefault && val === B3c?field[Z87]():val,h4R);if(!edit[x2E] || edit[b0u][name]){includeInOrder=X17;}}else {var Y6C=f9Cm$[555616];Y6C+=E7j;Y6C+=T7D;if(!edit[p6K] || edit[Y6C][name]){var f5e=C95;f5e+=T9t;f5e+=f9Cm$.J4L;field[f5e](idSrc,val === undefined || nullDefault && val === B3c?field[Z87]():val,h4R);includeInOrder=X17;}}}});field[H3$]();if(field[a4h]()[S1u] !== C37 && includeInOrder){usedFields[Z_J](name);}});var currOrder=this[U2b]()[o0_]();for(var i=currOrder[B97] - Y5Y;i >= C37;i--){var x7M=K18;x7M+=z8q;x7M+=g5f;if($[x7M](currOrder[i][a8d](),usedFields) === -Y5Y){var j4l=V1c;j4l+=N3e;currOrder[j4l](i,Y5Y);}}this[N$N](currOrder);this[G4T](p$6,[pluck(editFields,a1m)[C37],pluck(editFields,v$r)[C37],items,type],function(){var y3E='initMultiEdit';var z6X=j7g;z6X+=o6f;z6X+=f9Cm$.t_T;z6X+=l1E;I8Z.j$H();_this[z6X](y3E,[editFields,items,type],function(){setupDone();});});}function _event(trigger,args,promiseComplete){var s58="xOf";var m3y="Cance";var v1y='pre';var J3Q="esu";var d1D="lled";var y6X="rHandl";I8Z.j$H();if(args === void C37){args=[];}if(promiseComplete === void C37){promiseComplete=undefined;}if(Array[d2_](trigger)){for(var i=C37,ien=trigger[B97];i < ien;i++){this[G29](trigger[i],args);}}else {var E1_=E$W;E1_+=f9Cm$.t_T;E1_+=s58;var q6y=f9Cm$.l60;q6y+=J3Q;q6y+=Q6Q;q6y+=f9Cm$.J4L;var Z$d=K7O;Z$d+=f9Cm$.t_T;Z$d+=y6X;Z$d+=F7I;var e=$[A05](trigger);$(this)[Z$d](e,args);var result=e[q6y];if(trigger[E1_](v1y) === C37 && result === h4R){var e1p=m3y;e1p+=d1D;var P_4=i2S;P_4+=a6R;P_4+=f9Cm$[481343];P_4+=f9Cm$.J4L;$(this)[V39]($[P_4](trigger + e1p),args);}if(promiseComplete){var s6z=f9Cm$.J4L;s6z+=s$2;s6z+=f9Cm$.t_T;s6z+=f9Cm$[481343];if(result && typeof result === f9Cm$.c8L && result[s6z]){result[y$I](promiseComplete);}else {promiseComplete(result);}}return result;}}function _eventName(input){var W8y=/^on([A-Z])/;var D5w="substring";var W2U=3;var D1S="owerCas";var T6g="match";var Q0S=Q6Q;Q0S+=f9Cm$.t_T;Q0S+=f9Cm$[481343];I8Z.j$H();Q0S+=h5h;var E6_=V1c;E6_+=r0a;var name;var names=input[E6_](E$d);for(var i=C37,ien=names[Q0S];i < ien;i++){name=names[i];var onStyle=name[T6g](W8y);if(onStyle){var f_b=H$0;f_b+=n35;f_b+=D1S;f_b+=f9Cm$.t_T;name=onStyle[Y5Y][f_b]() + name[D5w](W2U);}names[i]=name;}return names[R5X](E$d);}function _fieldFromNode(node){var B8C=q1F;B8C+=X5O;var foundField=B3c;I8Z.j$H();$[a8N](this[f9Cm$.J3V][B8C],function(name,field){var e7P=f9Cm$[228782];e7P+=K18;e7P+=f9Cm$[481343];e7P+=f9Cm$[555616];var A0U=k14;A0U+=f9Cm$[555616];A0U+=f9Cm$.t_T;if($(field[A0U]())[e7P](node)[B97]){foundField=field;}});return foundField;}function _fieldNames(fieldNames){if(fieldNames === undefined){return this[P5P]();}else if(!Array[d2_](fieldNames)){return [fieldNames];}I8Z.m_d();return fieldNames;}function _focus(fieldsIn,focus){var r$N='div.DTE ';var S1H="nde";var J8T="activeEle";var Y6n="numb";var A_1=/^jq:/;var C5F=Y1m;C5F+=c_m;C5F+=J8D;C5F+=M42;var F_N=Y6n;F_N+=F7I;var _this=this;if(this[f9Cm$.J3V][Y8d] === m8Z){return;}var field;var fields=$[X5_](fieldsIn,function(fieldOrName){var z$q=F3o;z$q+=K5X;var l75=f9Cm$.J3V;l75+=J2x;l75+=f9Cm$[481343];l75+=h7H;return typeof fieldOrName === l75?_this[f9Cm$.J3V][z$q][fieldOrName]:fieldOrName;});if(typeof focus === F_N){field=fields[focus];}else if(focus){var M$3=f9Cm$.Z$r;M$3+=R2$;M$3+=I4e;var t2O=K18;t2O+=S1H;t2O+=G1H;t2O+=v0F;if(focus[t2O](M$3) === C37){var A$W=O4Y;A$W+=B1d;A$W+=i8m;field=$(r$N + focus[A$W](A_1,j_l));}else {field=this[f9Cm$.J3V][P5P][focus];}}else {var m6o=J0n;m6o+=Q6Q;m6o+=f9Cm$.Y87;m6o+=f9Cm$.l60;var H1_=J8T;H1_+=Z22;document[H1_][m6o]();}this[f9Cm$.J3V][C5F]=field;if(field){var Y8l=g2v;Y8l+=q9Z;Y8l+=M42;field[Y8l]();}}function _formOptions(opts){var t7S="closeI";var l7M="lea";var X8E="ssag";var J7l="oo";var M3h="canReturnSubmit";var e5O="efault";var o2k='.dteInline';var r73='keyup';var T7$="_fieldFromNode";var R2y="mes";var h32=t7S;h32+=q9Z;h32+=J0n;var g7O=i3K;g7O+=g5f;g7O+=f9Cm$[555616];g7O+=n2J;var t3O=f9Cm$[23424];t3O+=f9Cm$[481343];var q2c=J0n;q2c+=J7l;q2c+=l7M;q2c+=f9Cm$[481343];var W2_=a3D;W2_+=f9Cm$[481343];W2_+=f9Cm$.J3V;var p7C=e96;p7C+=j8w;p7C+=h8a;var l1S=f9Cm$.J4L;l1S+=g8Y;l1S+=k_d;var O_3=f9Cm$.J3V;O_3+=J2x;O_3+=f9Cm$[481343];O_3+=h7H;var Q_$=f9Cm$.J4L;Q_$+=K18;Q_$+=D3E;var _this=this;var that=this;var inlineCount=_inlineCounter++;var namespace=o2k + inlineCount;this[f9Cm$.J3V][x6t]=opts;this[f9Cm$.J3V][V53]=inlineCount;if(typeof opts[Q_$] === O_3 || typeof opts[l1S] === t0t){var G4J=f9Cm$.J4L;G4J+=g0r;G4J+=f9Cm$.t_T;this[M3D](opts[G4J]);opts[M3D]=X17;}if(typeof opts[c1_] === t3X || typeof opts[c1_] === p7C){var V49=W1R;V49+=M6l;var F6W=R2y;F6W+=f9Cm$.J3V;F6W+=D2Y;var k1A=W1R;k1A+=X8E;k1A+=f9Cm$.t_T;this[k1A](opts[F6W]);opts[V49]=X17;}if(typeof opts[W2_] !== q2c){var z8Q=J0n;z8Q+=g7i;z8Q+=f9Cm$.J4L;z8Q+=W4z;this[d7J](opts[d7J]);opts[z8Q]=X17;}$(document)[t3O](g7O + namespace,function(e){var T5w="canReturn";I8Z.m_d();var T7y="Submit";var q8O=H3b;q8O+=i8Z;q8O+=q9Z;q8O+=s$2;if(e[q8O] === m$Y && _this[f9Cm$.J3V][n5R]){var el=$(document[f$H]);if(el){var o3V=f9Cm$[228782];o3V+=e7h;o3V+=q5N;o3V+=f9Cm$[481343];var d4B=T5w;d4B+=T7y;var field=_this[T7$](el);if(field && typeof field[d4B] === o3V && field[M3h](el)){var q7P=C5p;q7P+=e5O;e[q7P]();}}}});$(document)[h8a](r73 + namespace,function(e){var H39="lur";var Q0K="prev";var a6Q="funct";var X2L="entD";var v_Z="nReturnS";var B6V="ents";var l2M="onReturn";var D05="etur";var J13="onEsc";var R6d="ive";var A0i=37;var d6Z="whic";var H9L=27;var Z1H="Elem";var E0s=39;var d90="eturn";var u_F="rev";var g2w='.DTE_Form_Buttons';var a$A=B1d;a$A+=f9Cm$.e08;a$A+=f9Cm$.l60;a$A+=B6V;var D4G=d6Z;D4G+=s$2;var S4h=u6E;S4h+=R6d;S4h+=Z1H;S4h+=u7S;var el=$(document[S4h]);if(e[D4G] === m$Y && _this[f9Cm$.J3V][n5R]){var o_9=a6Q;o_9+=X3e;var g5P=e7k;g5P+=v_Z;g5P+=T8h;var field=_this[T7$](el);if(field && typeof field[g5P] === o_9 && field[M3h](el)){var O6D=e96;O6D+=Q2A;O6D+=X3e;var F6G=h8a;F6G+=C_L;F6G+=D05;F6G+=f9Cm$[481343];var I2L=k1a;I2L+=f9Cm$.J4L;if(opts[l2M] === I2L){e[G8l]();_this[n3l]();}else if(typeof opts[F6G] === O6D){var Y50=f9Cm$[23424];Y50+=f9Cm$[481343];Y50+=C_L;Y50+=d90;var G9t=B1d;G9t+=u_F;G9t+=X2L;G9t+=e5O;e[G9t]();opts[Y50](_this,e);}}}else if(e[g4b] === H9L){var U0N=k18;U0N+=S_t;var d$5=q9Z;d$5+=C2Q;d$5+=f9Cm$.t_T;var B4P=J0n;B4P+=H39;var G8s=f9Cm$[228782];G8s+=e7h;G8s+=B4Y;G8s+=h8a;e[G8l]();if(typeof opts[J13] === G8s){var T0D=h8a;T0D+=i2S;T0D+=f9Cm$.J3V;T0D+=q9Z;opts[T0D](that,e);}else if(opts[J13] === B4P){var W9V=R6r;W9V+=f9Cm$.Y87;W9V+=f9Cm$.l60;that[W9V]();}else if(opts[J13] === d$5){var I3b=M4_;I3b+=Y1m;that[I3b]();}else if(opts[J13] === U0N){var I4R=f9Cm$.J3V;I4R+=q1I;I4R+=f9Cm$.J4L;that[I4R]();}}else if(el[a$A](g2w)[B97]){var i09=H3b;i09+=s$2;i09+=K18;i09+=j1L;if(e[g4b] === A0i){var k_F=q1i;k_F+=C_v;k_F+=f9Cm$[481343];el[Q0K](k_F)[c6V](F2a);}else if(e[i09] === E0s){var g7c=f9Cm$[228782];g7c+=f9Cm$[23424];g7c+=q9Z;g7c+=M42;var K4H=E2j;K4H+=f9Cm$.J4L;K4H+=f9Cm$[23424];K4H+=f9Cm$[481343];var f_C=f9Cm$[481343];f_C+=f9Cm$.t_T;f_C+=G1H;f_C+=f9Cm$.J4L;el[f_C](K4H)[c6V](g7c);}}});this[f9Cm$.J3V][h32]=function(){var U7p=i3K;U7p+=g5f;U7p+=f9Cm$.Y87;U7p+=B1d;var F2Q=H_E;F2Q+=f9Cm$[228782];I8Z.m_d();$(document)[v93](w_k + namespace);$(document)[F2Q](U7p + namespace);};return namespace;}function _inline(editFields,opts,closeCb){var r3J="xO";var a6W='cancel';var I7q='.';var c6X="mError";var e3H="styl";var j2M="butt";var X4C="ine";var N6s="userAgent";var l9i="_inputTr";var k0G="ass=\"";var u1Z="engt";var F8J="iv cl";var e64="tach";var b2c="iv";var W_5="e=\"width:";var I0A="gger";var I5N="asses";var p3$="v.";var V0n="inl";var v83="_preo";var v3Y='<div class="DTE_Processing_Indicator"><span></span></div>';var A8Z="ild";var L_f="contents";var A9X=f9Cm$[228782];A9X+=f9Cm$[23424];A9X+=q9Z;I8Z.m_d();A9X+=M42;var I7u=l9i;I7u+=K18;I7u+=I0A;var r0r=N5j;r0r+=f2C;r0r+=f9Cm$.J4L;var Q2W=B96;Q2W+=k1J;Q2W+=f9Cm$[481343];Q2W+=f9Cm$.t_T;var b4Z=v83;b4Z+=V7f;var C1G=Q6Q;C1G+=u1Z;C1G+=s$2;var I1r=e3G;I1r+=f9Cm$.t_T;I1r+=g5f;I1r+=f9Cm$.J3V;var H2L=V0n;H2L+=X4C;var I5J=q9Z;I5J+=Q6Q;I5J+=I5N;var _this=this;if(closeCb === void C37){closeCb=B3c;}var closed=h4R;var classes=this[I5J][H2L];var keys=Object[I1r](editFields);var editRow=editFields[keys[C37]];var lastAttachPoint;var elements=[];for(var i=C37;i < editRow[P8n][C1G];i++){var T$K=k6$;T$K+=e64;var o5y=f9Cm$[228782];o5y+=K18;o5y+=G6S;o5y+=X5O;var g1G=B1d;g1G+=M42;g1G+=s$2;var h8V=P8n;h8V+=y7j;h8V+=e6e;var name_1=editRow[h8V][i][C37];elements[g1G]({field:this[f9Cm$.J3V][o5y][name_1],name:name_1,node:$(editRow[T$K][i])});}var namespace=this[Q49](opts);var ret=this[b4Z](Q2W);if(!ret){return this;}for(var _i=C37,elements_1=elements;_i < elements_1[B97];_i++){var c_I=f9Cm$[481343];c_I+=f9Cm$[23424];c_I+=f9Cm$[555616];c_I+=f9Cm$.t_T;var D_y=f9Cm$[228782];D_y+=K18;D_y+=D9y;var S4k=i$c;S4k+=c6X;var V98=o9d;V98+=D6p;var n1o=C$g;n1o+=p7O;n1o+=T52;var Z26=f9Cm$[481343];Z26+=f9Cm$[23424];Z26+=R3S;var t0i=k1J;t0i+=I$m;var T8o=s4A;T8o+=T5B;var M8p=k8A;M8p+=f9Cm$[555616];M8p+=F8J;M8p+=k0G;var c50=o0P;c50+=b2c;c50+=i$Z;var Q5x=y9e;Q5x+=f_l;var N8R=Q6Q;N8R+=K18;N8R+=L$V;N8R+=f9Cm$.l60;var e1V=B1d;e1V+=G1H;e1V+=y9e;var J6C=e3H;J6C+=W_5;var i_h=C5j;i_h+=P9U;i_h+=R$r;var R32=E$W;R32+=f9Cm$.t_T;R32+=r3J;R32+=f9Cm$[228782];var B9h=R3S;B9h+=e64;var F9t=j1L;F9t+=A8Z;F9t+=w2Q;var s1A=f9Cm$[481343];s1A+=f9Cm$[23424];s1A+=f9Cm$[555616];s1A+=f9Cm$.t_T;var el=elements_1[_i];var node=el[s1A];el[F9t]=node[L_f]()[B9h]();var style=navigator[N6s][R32](i_h) !== -Y5Y?J6C + node[T4U]() + e1V:j_l;node[P2$]($(v4Q + classes[A2A] + x$l + v4Q + classes[N8R] + Q5x + style + f$B + v3Y + c50 + M8p + classes[d7J] + T8o + Q4m));node[I$n](v1_ + classes[t0i][d9I](/ /g,I7q))[P2$](el[D_P][Z26]())[n1o](this[V98][S4k]);lastAttachPoint=el[D_y][c_I]();if(opts[d7J]){var O_D=q1i;O_D+=I4O;O_D+=f9Cm$[23424];O_D+=f9c;var b8c=j2M;b8c+=h8a;b8c+=f9Cm$.J3V;var G8p=h44;G8p+=p3$;node[I$n](G8p + classes[b8c][d9I](/ /g,I7q))[P2$](this[X6t][O_D]);}}var submitClose=this[I45](r0r,opts,lastAttachPoint);var cancelClose=this[I7u](a6W,opts,lastAttachPoint);this[k7m](function(submitComplete,action){var x38="inli";var E1D="forEach";var U6W=x38;U6W+=L$V;var Z0c=f9Cm$.t_T;Z0c+=f9Cm$[555616];Z0c+=K18;Z0c+=f9Cm$.J4L;var A5q=q9Z;A5q+=k1J;A5q+=K2u;closed=X17;$(document)[v93](A5q + namespace);if(!submitComplete || action !== Z0c){elements[E1D](function(el){var N2f=C$g;N2f+=G7H;var c3V=k14;c3V+=f9Cm$[555616];c3V+=f9Cm$.t_T;var H_4=s_E;H_4+=j1L;I8Z.j$H();var M_U=Y$v;M_U+=u7S;M_U+=f9Cm$.J3V;var B5c=k14;B5c+=f9Cm$[555616];B5c+=f9Cm$.t_T;el[B5c][M_U]()[H_4]();el[c3V][N2f](el[n_U]);});}submitClose();cancelClose();_this[K5s]();if(closeCb){closeCb();}return U6W;;});setTimeout(function(){var I4i="usedo";var X4c='andSelf';var v8m=M_R;v8m+=K18;v8m+=q9Z;v8m+=e3G;var U6v=u_A;I8Z.j$H();U6v+=I4i;U6v+=H3b;U6v+=f9Cm$[481343];var t3J=O1b;t3J+=f9Cm$[555616];t3J+=k7Z;if(closed){return;}var back=$[f9Cm$.E4X][j87]?t3J:X4c;var target;$(document)[h8a](U6v + namespace,function(e){var D_m="ar";I8Z.m_d();var M$2=f9Cm$.J4L;M$2+=D_m;M$2+=P9U;M$2+=f9Cm$.J4L;target=e[M$2];})[h8a](w_k + namespace,function(e){var G6M="rge";var t9h=f9Cm$.J4L;t9h+=f9Cm$.e08;t9h+=G6M;t9h+=f9Cm$.J4L;target=e[t9h];})[h8a](v8m + namespace,function(e){var Q63='owns';var isIn=h4R;for(var _i=C37,elements_2=elements;_i < elements_2[B97];_i++){var g0A=f9Cm$[481343];g0A+=x4Y;var P8T=K18;P8T+=i0Z;P8T+=q7Q;P8T+=g5f;var el=elements_2[_i];if(el[D_P][Y_T](Q63,target) || $[P8T](el[g0A][C37],$(target)[z89]()[back]()) !== -Y5Y){isIn=X17;}}if(!isIn){_this[h1_]();}});},C37);this[U5L]($[X5_](elements,function(el){var V3v=q1F;I8Z.j$H();V3v+=f9Cm$[555616];return el[V3v];}),opts[A9X]);this[s3c](O8g,X17);}function _inputTrigger(type,opts,insertPoint){var n7U='click.dte-';var F4b="childNode";var l5p="ldre";I8Z.j$H();var c6C='number';var H88="Trigge";var R5a="closest";var D5l='Html';var R85=F4b;R85+=f9Cm$.J3V;var B4V=f9Cm$.J4L;B4V+=f9Cm$.l60;var c$G=H88;c$G+=f9Cm$.l60;var _this=this;var trigger=opts[type + c$G];var html=opts[type + D5l];var event=n7U + type;var tr=$(insertPoint)[R5a](B4V);if(trigger === undefined){return function(){};}if(typeof trigger === c6C){var J6x=Q6Q;J6x+=n3_;J6x+=i1I;J6x+=s$2;var i4l=q9Z;i4l+=i8Z;i4l+=l5p;i4l+=f9Cm$[481343];var kids=tr[i4l]();trigger=trigger < C37?kids[kids[J6x] + trigger]:kids[trigger];}var children=$(trigger,tr)[B97]?Array[P99][k3q][B7t]($(trigger,tr)[C37][R85]):[];$(children)[Z3_]();var triggerEl=$(trigger,tr)[h8a](event,function(e){I8Z.m_d();var z4P=c18;z4P+=f9Cm$.t_T;z4P+=Q6Q;e[T7J]();if(type === z4P){_this[H9B]();}else {_this[n3l]();}})[P2$](html);return function(){var g4B="pty";var Y_C=A1Z;I8Z.m_d();Y_C+=g4B;triggerEl[v93](event)[Y_C]()[P2$](children);};}function _optionsUpdate(json){var V6M=W6q;I8Z.j$H();V6M+=n59;V6M+=f9c;var that=this;if(json && json[V6M]){var Z6f=f9Cm$[228782];Z6f+=f$A;Z6f+=X5O;var R65=f9Cm$.t_T;R65+=f9Cm$.e08;R65+=q9Z;R65+=s$2;$[R65](this[f9Cm$.J3V][Z6f],function(name,field){var f2F="updat";if(json[O$K][name] !== undefined){var fieldInst=that[D_P](name);if(fieldInst && fieldInst[T0s]){var i2H=O11;i2H+=h8a;i2H+=f9Cm$.J3V;var c4g=f2F;c4g+=f9Cm$.t_T;fieldInst[c4g](json[i2H][name]);}}});}}function _message(el,msg,title,fn){var g9v="eO";var z4e="eAt";var B3y="deI";var x2v=f9Cm$[228782];x2v+=f9Cm$[481343];var canAnimate=$[x2v][K5h]?X17:h4R;if(title === undefined){title=h4R;}if(!fn){fn=function(){};}if(typeof msg === t0t){msg=msg(this,new DataTable$4[s5A](this[f9Cm$.J3V][z6u]));}I8Z.j$H();el=$(el);if(canAnimate){el[Y4v]();}if(!msg){var D3Y=C1Z;D3Y+=B1d;D3Y+=I8$;D3Y+=w3$;if(this[f9Cm$.J3V][D3Y] && canAnimate){var H3X=f9Cm$[228782];H3X+=O1b;H3X+=g9v;H3X+=g7i;el[H3X](function(){el[n8n](j_l);I8Z.j$H();fn();});}else {var T$H=s$2;T$H+=f9Cm$.J4L;T$H+=D6p;T$H+=Q6Q;el[T$H](j_l)[X5f](M_p,n50);fn();}if(title){var Q5L=O4Y;Q5L+=U7W;Q5L+=z4e;Q5L+=Y7m;el[Q5L](f7V);}}else {fn();if(this[f9Cm$.J3V][n5R] && canAnimate){var g6u=f9Cm$[228782];g6u+=f9Cm$.e08;g6u+=B3y;g6u+=f9Cm$[481343];el[n8n](msg)[g6u]();}else {var E3A=h44;E3A+=t_U;E3A+=f9Cm$.e08;E3A+=g5f;el[n8n](msg)[X5f](E3A,M20);}if(title){var s8Q=f9Cm$.e08;s8Q+=f9Cm$.J4L;s8Q+=f9Cm$.J4L;s8Q+=f9Cm$.l60;el[s8Q](f7V,msg);}}}function _multiInfo(){var O96="Mul";var J2H="isMultiVal";var C1W="hown";var j4D="oS";var w_t="includeFiel";var M1c="tiValue";var n0B=e2b;n0B+=K5v;var G9D=w_t;G9D+=X5O;var K5C=F3o;K5C+=f9Cm$.t_T;K5C+=o$D;var fields=this[f9Cm$.J3V][K5C];var include=this[f9Cm$.J3V][G9D];var show=X17;var state;if(!include){return;}for(var i=C37,ien=include[n0B];i < ien;i++){var c4v=I_i;c4v+=j4D;c4v+=C1W;var V4i=J2H;V4i+=X3P;var P9s=K_b;P9s+=O96;P9s+=M1c;var P5g=M6F;P5g+=e$q;P5g+=x1X;var field=fields[include[i]];var multiEditable=field[P5g]();if(field[P9s]() && multiEditable && show){state=X17;show=h4R;}else if(field[V4i]() && !multiEditable){state=X17;}else {state=h4R;}fields[include[i]][c4v](state);}}function _nestedClose(cb){var g_f="roller";var t95="displayCont";var o8D="callback";var P0z=Q6Q;P0z+=d3h;var a_j=k_d;a_j+=L47;a_j+=K5v;var W95=j7g;W95+=f9Cm$.J3V;W95+=u1J;W95+=H3b;var disCtrl=this[f9Cm$.J3V][I2G];var show=disCtrl[W95];if(!show || !show[a_j]){if(cb){cb();}}else if(show[P0z] > Y5Y){var N1k=D_8;N1k+=f9Cm$[555616];var X6g=f9Cm$[555616];X6g+=f9Cm$.J4L;X6g+=f9Cm$.t_T;var u13=f9Cm$[23424];u13+=B1d;u13+=f9Cm$.t_T;u13+=f9Cm$[481343];var s1I=t95;s1I+=g_f;var B1n=Q6Q;B1n+=f9Cm$.t_T;B1n+=L47;B1n+=K5v;var Z1b=B1d;Z1b+=J_q;show[Z1b]();var last=show[show[B1n] - Y5Y];if(cb){cb();}this[f9Cm$.J3V][s1I][u13](last[X6g],last[N1k],last[o8D]);}else {var U0O=e2b;U0O+=K5v;var q66=M_R;q66+=J_E;this[f9Cm$.J3V][I2G][q66](this,cb);show[U0O]=C37;}}function _nestedOpen(cb,nest){var R5s="_sh";var I1X="_show";var A6X="isplayController";var Z1L=f9Cm$[23424];Z1L+=B1d;Z1L+=f9Cm$.t_T;Z1L+=f9Cm$[481343];var q2$=p5y;q2$+=f9Cm$.e08;q2$+=B1d;q2$+=o6H;var x3d=f9Cm$[555616];x3d+=f9Cm$[23424];x3d+=D6p;var c$b=B1d;c$b+=f9Cm$.Y87;c$b+=f9Cm$.J3V;c$b+=s$2;var Z1v=R5s;I8Z.m_d();Z1v+=h98;var S4c=f9Cm$[555616];S4c+=A6X;var disCtrl=this[f9Cm$.J3V][S4c];if(!disCtrl[I1X]){disCtrl[I1X]=[];}if(!nest){disCtrl[I1X][B97]=C37;}disCtrl[Z1v][c$b]({append:this[x3d][q2$],callback:cb,dte:this});this[f9Cm$.J3V][I2G][Z1L](this,this[X6t][A2A],cb);}function _postopen(type,immediate){var A5u='submit.editor-internal';var Q$X=".editor";var w7R="-inter";var z49="_eve";var S5D="playContr";var S9Y="oller";var Y0S="nal";var v$J="-focus";var T3R="captureFocus";var d$I="submit.ed";var w8Y=f9Cm$[23424];w8Y+=p7O;w8Y+=f9Cm$[481343];var w5O=z49;w5O+=l1E;var t96=D6p;t96+=f9Cm$.e08;t96+=K18;t96+=f9Cm$[481343];var j91=f9Cm$[23424];j91+=f9Cm$[481343];var h7a=d$I;h7a+=N7v;h7a+=w7R;h7a+=Y0S;var N$e=i$c;N$e+=D6p;var y2b=f9Cm$[555616];y2b+=f9Cm$[23424];y2b+=D6p;var T8R=f9Cm$[555616];T8R+=K_b;T8R+=S5D;T8R+=S9Y;var _this=this;I8Z.j$H();var focusCapture=this[f9Cm$.J3V][T8R][T3R];if(focusCapture === undefined){focusCapture=X17;}$(this[y2b][N$e])[v93](h7a)[j91](A5u,function(e){var Z5O="efa";var E_t=C5p;E_t+=Z5O;E_t+=F8_;e[E_t]();});if(focusCapture && (type === t96 || type === t0m)){var c8P=G0n;c8P+=Q$X;c8P+=v$J;var w6Z=f9Cm$[23424];w6Z+=f9Cm$[481343];var e1X=R$W;e1X+=g5f;$(e1X)[w6Z](c8P,function(){var Z48="TED";var R45='.DTE';var c0$="arent";var y2l="active";var N$G="tFo";var n5U="Elemen";var N9k=Q6Q;N9k+=n3_;N9k+=i1I;N9k+=s$2;var V8N=u5z;V8N+=l93;V8N+=Z48;var d91=B1d;d91+=c0$;d91+=f9Cm$.J3V;var e2i=Q6Q;e2i+=q8V;e2i+=f9Cm$.J4L;e2i+=s$2;var W3S=y2l;W3S+=n5U;I8Z.j$H();W3S+=f9Cm$.J4L;if($(document[W3S])[z89](R45)[e2i] === C37 && $(document[f$H])[d91](V8N)[N9k] === C37){var r9f=Y1m;r9f+=c_m;r9f+=J8D;r9f+=M42;if(_this[f9Cm$.J3V][r9f]){var v0E=d0M;v0E+=M42;var O2r=Y1m;O2r+=N$G;O2r+=e4v;O2r+=f9Cm$.J3V;_this[f9Cm$.J3V][O2r][v0E]();}}});}this[m_q]();this[w5O](w8Y,[type,this[f9Cm$.J3V][Y8d]]);if(immediate){var v8l=f9Cm$.e08;v8l+=q9Z;v8l+=q5N;v8l+=f9Cm$[481343];var S38=j7g;S38+=o6f;S38+=f9Cm$.t_T;S38+=l1E;this[S38](I27,[type,this[f9Cm$.J3V][v8l]]);}return X17;}function _preopen(type){var T75="eve";var o63="cb";var S0t="eOpe";var X9U="aye";var d5g="cancelO";var V9D="seI";var f2p=s8o;f2p+=X9U;f2p+=f9Cm$[555616];var g16=u6g;g16+=S0t;g16+=f9Cm$[481343];var a0C=j7g;a0C+=T75;a0C+=l1E;if(this[a0C](g16,[type,this[f9Cm$.J3V][Y8d]]) === h4R){var H$X=C_B;H$X+=k_d;var W$k=D6p;W$k+=f9Cm$[23424];W$k+=f9Cm$[555616];W$k+=f9Cm$.t_T;var D3y=u6E;D3y+=X3e;var n$n=d5g;n$n+=V7f;this[K5s]();this[G29](n$n,[type,this[f9Cm$.J3V][D3y]]);if((this[f9Cm$.J3V][h56] === O8g || this[f9Cm$.J3V][W$k] === H$X) && this[f9Cm$.J3V][c78]){var y2r=M_R;y2r+=f9Cm$[23424];y2r+=V9D;y2r+=o63;this[f9Cm$.J3V][y2r]();}this[f9Cm$.J3V][c78]=B3c;return h4R;}this[K5s](X17);this[f9Cm$.J3V][f2p]=type;I8Z.j$H();return X17;}function _processing(processing){var b27="wra";var u_j="toggleClass";var p7k="roce";var y9u=G2J;y9u+=x_L;y9u+=B96;y9u+=h7H;var D_s=b27;D_s+=y7w;var l8z=B2B;l8z+=f2c;var y9B=o$n;I8Z.j$H();y9B+=B4Y;y9B+=I7r;y9B+=f9Cm$.t_T;var k7I=B1d;k7I+=p7k;k7I+=f9Cm$.J3V;k7I+=G7W;var procClass=this[A2x][k7I][y9B];$([l8z,this[X6t][D_s]])[u_j](procClass,processing);this[f9Cm$.J3V][y9u]=processing;this[G29](H27,[processing]);}function _noProcessing(args){var Q0p="processing-fiel";var processing=h4R;I8Z.j$H();$[a8N](this[f9Cm$.J3V][P5P],function(name,field){if(field[T0W]()){processing=X17;}});if(processing){var e37=Q0p;e37+=f9Cm$[555616];var i1_=f9Cm$[23424];i1_+=f9Cm$[481343];i1_+=f9Cm$.t_T;this[i1_](e37,function(){var k0a="sin";var M0G="_submit";var G8v="_noProces";var h7D=G8v;h7D+=k0a;h7D+=h7H;if(this[h7D](args) === X17){this[M0G][D08](this,args);}});}return !processing;}function _submit(successCallback,errorCallback,formatdata,hide){var H34='Field is still processing';var x4r="ha";var m$w='allIfChanged';var w2B="tD";var m9k="_clos";var k7k="actionName";var d59="tOpts";var G2R="functi";var B5m="nge";var d2E=16;var K$m=B1d;K$m+=K$K;K$m+=J0n;I8Z.m_d();K$m+=S_t;var d8p=O0W;d8p+=o3Z;var u21=f9Cm$.t_T;u21+=f9Cm$[555616];u21+=K18;u21+=f9Cm$.J4L;var b3l=f9Cm$.t_T;b3l+=h44;b3l+=d59;var q1t=G3n;q1t+=w2B;q1t+=k6$;q1t+=f9Cm$.e08;var l8F=F3o;l8F+=f9Cm$.t_T;l8F+=Q6Q;l8F+=X5O;var _this=this;var changed=h4R;var allData={};var changedData={};var setBuilder=dataSet;var fields=this[f9Cm$.J3V][l8F];var editCount=this[f9Cm$.J3V][V53];var editFields=this[f9Cm$.J3V][t67];var editData=this[f9Cm$.J3V][q1t];var opts=this[f9Cm$.J3V][b3l];var changedSubmit=opts[n3l];var submitParamsLocal;if(this[S3a](arguments) === h4R){var p3z=f9Cm$.t_T;p3z+=C58;p3z+=A5h;Editor[p3z](H34,d2E,h4R);return;}var action=this[f9Cm$.J3V][Y8d];var submitParams={data:{}};submitParams[this[f9Cm$.J3V][k7k]]=action;if(action === c2D || action === u21){var H06=q9Z;H06+=x4r;H06+=B5m;H06+=f9Cm$[555616];$[a8N](editFields,function(idSrc,edit){var allRowData={};var changedRowData={};$[a8N](fields,function(name,field){var X$z="ndexOf";var Q1$="comp";var e54="y-co";var V1W="submittable";var M1p="valFrom";var E5E="man";var O1_="are";var A24="unt";var m9R=/\[.*$/;var c4y="[";var o7o="-";if(edit[P5P][name] && field[V1W]()){var Y7H=Q1$;Y7H+=O1_;var y37=o7o;y37+=E5E;y37+=e54;y37+=A24;var H7n=f9Cm$.l60;H7n+=d9i;H7n+=f9Cm$.e08;H7n+=z8x;var v8T=c4y;v8T+=M3E;var y$C=K18;y$C+=X$z;var C7L=f9Cm$.J3V;C7L+=J2x;C7L+=L47;var s7N=K_b;s7N+=W0c;s7N+=V9A;var multiGet=field[o1D]();var builder=setBuilder(name);if(multiGet[idSrc] === undefined){var T5P=f9Cm$[555616];T5P+=f9Cm$.e08;T5P+=Q3e;var N85=M1p;N85+=l93;N85+=q94;var originalVal=field[N85](edit[T5P]);builder(allRowData,originalVal);return;}var value=multiGet[idSrc];var manyBuilder=Array[s7N](value) && typeof name === C7L && name[y$C](v8T) !== -Y5Y?setBuilder(name[H7n](m9R,j_l) + y37):B3c;builder(allRowData,value);if(manyBuilder){var D7X=D3W;D7X+=h5h;manyBuilder(allRowData,value[D7X]);}if(action === Z75 && (!editData[name] || !field[Y7H](value,editData[name][idSrc]))){builder(changedRowData,value);changed=X17;if(manyBuilder){manyBuilder(changedRowData,value[B97]);}}}});if(!$[E8u](allRowData)){allData[idSrc]=allRowData;}I8Z.m_d();if(!$[E8u](changedRowData)){changedData[idSrc]=changedRowData;}});if(action === c2D || changedSubmit === D7e || changedSubmit === m$w && changed){var Y_0=f9Cm$[555616];Y_0+=f9Cm$.e08;Y_0+=Q3e;submitParams[Y_0]=allData;}else if(changedSubmit === H06 && changed){var N8P=f9Cm$[555616];N8P+=q94;submitParams[N8P]=changedData;}else {var K44=G2R;K44+=f9Cm$[23424];K44+=f9Cm$[481343];var R4X=v1S;R4X+=f9Cm$.J4L;R4X+=f9Cm$.t_T;var o15=M_R;o15+=i$r;o15+=f9Cm$.t_T;var Q68=f9Cm$.e08;Q68+=j8w;Q68+=f9Cm$[23424];Q68+=f9Cm$[481343];this[f9Cm$.J3V][Q68]=B3c;if(opts[p0i] === o15 && (hide === undefined || hide)){var C2K=m9k;C2K+=f9Cm$.t_T;this[C2K](h4R);}else if(typeof opts[R4X] === K44){opts[p0i](this);}if(successCallback){successCallback[B7t](this);}this[I0J](h4R);this[G29](L5$);return;}}else if(action === d8p){$[a8N](editFields,function(idSrc,edit){var G1w=f9Cm$[555616];G1w+=f9Cm$.e08;G1w+=f9Cm$.J4L;G1w+=f9Cm$.e08;submitParams[G1w][idSrc]=edit[I$4];});}submitParamsLocal=$[d7H](X17,{},submitParams);if(formatdata){formatdata(submitParams);}this[G29](K$m,[submitParams,action],function(result){var D$$="rocessin";var p_$="submitTabl";I8Z.m_d();var K4T="_aj";if(result === h4R){var v5i=q_Z;v5i+=D$$;v5i+=h7H;_this[v5i](h4R);}else {var m7b=q9Z;m7b+=M_7;m7b+=Q6Q;var N$V=j7g;N$V+=p_$;N$V+=f9Cm$.t_T;var l5b=K4T;l5b+=J_Y;var y_K=f9Cm$.e08;y_K+=Q23;y_K+=G1H;var submitWire=_this[f9Cm$.J3V][y_K]?_this[l5b]:_this[N$V];submitWire[m7b](_this,submitParams,function(json,notGood,xhr){I8Z.m_d();var c5o="_submitSuccess";_this[c5o](json,notGood,submitParams,submitParamsLocal,_this[f9Cm$.J3V][Y8d],editCount,hide,successCallback,errorCallback,xhr);},function(xhr,err,thrown){I8Z.m_d();_this[C4S](xhr,err,thrown,errorCallback,submitParams,_this[f9Cm$.J3V][Y8d]);},submitParams);}});}function _submitTable(data,success,error,submitParams){var M62="dataSou";var z7Y="ivid";var p4s="ual";var g7e=f9Cm$.l60;g7e+=f9Cm$.t_T;g7e+=u_A;g7e+=a6R;var L6I=K18;L6I+=C$G;L6I+=q9Z;var action=data[Y8d];var out={data:[]};var idGet=dataGet(this[f9Cm$.J3V][L6I]);var idSet=dataSet(this[f9Cm$.J3V][c5v]);if(action !== g7e){var E13=f9Cm$.t_T;E13+=f9Cm$.e08;E13+=q9Z;E13+=s$2;var D4D=E$W;D4D+=z7Y;D4D+=p4s;var r_2=q1F;r_2+=X5O;var E6q=j7g;E6q+=M62;E6q+=j21;var f7p=D6p;f7p+=f9Cm$.e08;f7p+=K18;f7p+=f9Cm$[481343];var y5z=L$j;y5z+=f9Cm$.t_T;var originalData_1=this[f9Cm$.J3V][y5z] === f7p?this[E6q](r_2,this[n9R]()):this[e3V](D4D,this[n9R]());$[E13](data[I$4],function(key,vals){var R7h=g$w;R7h+=f9Cm$.J3V;R7h+=s$2;var toSave;var extender=extend;if(action === Z75){var U5J=f9Cm$[555616];U5J+=f9Cm$.e08;U5J+=f9Cm$.J4L;U5J+=f9Cm$.e08;var rowData=originalData_1[key][U5J];toSave=extender({},rowData,X17);toSave=extender(toSave,vals,X17);}else {toSave=extender({},vals,X17);}var overrideId=idGet(toSave);if(action === c2D && overrideId === undefined){idSet(toSave,+new Date() + key[a8d]());}else {idSet(toSave,overrideId);}out[I$4][R7h](toSave);});}success(out);}function _submitSuccess(json,notGood,submitParams,submitParamsLocal,action,editCount,hide,successCallback,errorCallback,xhr){var N9D='submitSuccess';var F5H='postCreate';var T8K="_dataSo";var k4Q="crea";var R8h='submitUnsuccessful';var M$c="aSou";var f82="editCou";var B0j="eEdit";I8Z.j$H();var W78="ubm";var f5T="Sourc";var o0v="tS";var G1U="preC";var r_k="fieldErro";var m_D="modifi";var p4o='preRemove';var J9Y="commi";var k6n="rrors";var i9q='setData';var u$_='commit';var l6g='prep';var Z3H='postEdit';var a5E="ors";var Z1S="dErr";var m1Z="fieldE";var h1O="dE";var s_x="onCompl";var f2$="postRemo";var h_p=p3g;h_p+=I7r;h_p+=f9Cm$.t_T;h_p+=l1E;var y1h=D3W;y1h+=h5h;var e3N=s4L;e3N+=Q6Q;e3N+=Z1S;e3N+=a5E;var n9V=t3Z;n9V+=f9Cm$.l60;var w9x=m1Z;w9x+=k6n;var q90=y3N;q90+=A5h;var l9v=N8o;l9v+=o0v;l9v+=W78;l9v+=g8Y;var G98=j7g;G98+=f9Cm$.t_T;G98+=e2C;var e9c=m_D;e9c+=F7I;var _this=this;var that=this;var setData;var fields=this[f9Cm$.J3V][P5P];var opts=this[f9Cm$.J3V][x6t];var modifier=this[f9Cm$.J3V][e9c];this[G98](l9v,[json,submitParams,action,xhr]);if(!json[q90]){json[h8G]=j_l;}if(!json[w9x]){var i$C=F3o;i$C+=G6S;i$C+=h1O;i$C+=k6n;json[i$C]=[];}if(notGood || json[n9V] || json[e3N][y1h]){var B7E=p3g;B7E+=e2C;var L4T=k8A;L4T+=J0n;L4T+=f9Cm$.l60;L4T+=i$Z;var v$L=y3N;v$L+=A5h;var E7B=r_k;E7B+=h_u;var globalError_1=[];if(json[h8G]){globalError_1[Z_J](json[h8G]);}$[a8N](json[E7B],function(i,err){var f1d="siti";var t5J="func";var y20='Unknown field: ';var t_3='Error';var k7Y="onFieldError";var c_h="dyCon";var K1a="po";var x$H="cus";var field=fields[err[h2d]];if(!field){throw new Error(y20 + err[h2d]);}else if(field[n5R]()){var f$C=i2S;f$C+=f9Cm$.l60;f$C+=a_T;f$C+=f9Cm$.l60;field[h8G](err[C8k] || f$C);if(i === C37){var q3f=t5J;q3f+=O74;var T4M=f9Cm$[228782];T4M+=f9Cm$[23424];T4M+=x$H;if(opts[k7Y] === T4M){var H62=K1a;H62+=f1d;H62+=h8a;var H1r=k14;H1r+=R3S;var k9o=r0K;k9o+=c_h;k9o+=f9Cm$.J4L;k9o+=u7S;_this[T4f]($(_this[X6t][k9o]),{scrollTop:$(field[H1r]())[H62]()[j52]},J0Z);field[G0n]();}else if(typeof opts[k7Y] === q3f){opts[k7Y](_this,err);}}}else {var H6m=I4e;H6m+=f_l;globalError_1[Z_J](field[h2d]() + H6m + (err[C8k] || t_3));}});this[v$L](globalError_1[R5X](L4T));this[B7E](R8h,[json]);if(errorCallback){errorCallback[B7t](that,json);}}else {var B9n=f82;B9n+=l1E;var R8q=k4Q;R8q+=U58;var store={};if(json[I$4] && (action === R8q || action === Z75)){var k4C=f9Cm$[555616];k4C+=f9Cm$.e08;k4C+=f9Cm$.J4L;k4C+=f9Cm$.e08;var v2L=J9Y;v2L+=f9Cm$.J4L;var b$D=T8K;b$D+=P7B;this[e3V](l6g,action,modifier,submitParamsLocal,json,store);for(var _i=C37,_a=json[I$4];_i < _a[B97];_i++){var c$A=f9Cm$.t_T;c$A+=f9Cm$[555616];c$A+=K18;c$A+=f9Cm$.J4L;var B0A=q9Z;B0A+=g7G;B0A+=f9Cm$.J4L;B0A+=f9Cm$.t_T;var W68=p3g;W68+=I7r;W68+=n3_;W68+=f9Cm$.J4L;var T6k=K18;T6k+=f9Cm$[555616];var data=_a[_i];setData=data;var id=this[e3V](T6k,data);this[W68](i9q,[json,data,action]);if(action === B0A){var g6N=q9Z;g6N+=f9Cm$.l60;g6N+=n_a;var s7F=G1U;s7F+=g7G;s7F+=f9Cm$.J4L;s7F+=f9Cm$.t_T;this[G29](s7F,[json,data,id]);this[e3V](c2D,fields,data,store);this[G29]([g6N,F5H],[json,data,id]);}else if(action === c$A){var X2W=f9Cm$.t_T;X2W+=D9$;var S9G=f9Cm$.t_T;S9G+=f9Cm$[555616];S9G+=g8Y;var y3v=B1d;y3v+=f9Cm$.l60;y3v+=B0j;this[G29](y3v,[json,data,id]);this[e3V](S9G,modifier,fields,data,store);this[G29]([X2W,Z3H],[json,data,id]);}}this[b$D](v2L,action,modifier,json[k4C],store);}else if(action === m8Z){var d5R=I_j;d5R+=k6$;d5R+=M$c;d5R+=j21;var G7M=K18;G7M+=f9Cm$[555616];G7M+=f9Cm$.J3V;var e0I=f2$;e0I+=a6R;var F9Y=f9Cm$.l60;F9Y+=Y3n;var Q0I=I_j;Q0I+=q94;Q0I+=f5T;Q0I+=f9Cm$.t_T;var y0z=K18;y0z+=f9Cm$[555616];y0z+=f9Cm$.J3V;var r2k=j7g;r2k+=n_A;var x7f=B1d;x7f+=f9Cm$.l60;x7f+=f9Cm$.t_T;x7f+=B1d;this[e3V](x7f,action,modifier,submitParamsLocal,json,store);this[r2k](p4o,[json,this[y0z]()]);this[Q0I](F9Y,modifier,fields,store);this[G29]([m8Z,e0I],[json,this[G7M]()]);this[d5R](u$_,action,modifier,json[I$4],store);}if(editCount === this[f9Cm$.J3V][B9n]){var b4I=q9Z;b4I+=h_t;var a30=s_x;a30+=s67;var h7M=o$n;h7M+=B4Y;h7M+=h8a;var sAction=this[f9Cm$.J3V][h7M];this[f9Cm$.J3V][Y8d]=B3c;if(opts[a30] === b4I && (hide === undefined || hide)){var A8f=j7g;A8f+=q9Z;A8f+=C2Q;A8f+=f9Cm$.t_T;this[A8f](json[I$4]?X17:h4R,sAction);}else if(typeof opts[p0i] === t0t){var S_h=v1S;S_h+=U58;opts[S_h](this);}}if(successCallback){var q35=q9Z;q35+=f9Cm$.e08;q35+=Q6Q;q35+=Q6Q;successCallback[q35](that,json);}this[G29](N9D,[json,setData,action]);}this[I0J](h4R);this[h_p](L5$,[json,setData,action]);}function _submitError(xhr,err,thrown,errorCallback,submitParams,action){var c4K='postSubmit';var i49='submitError';var c0r="syst";var F4N="itComplet";var h4$=C9K;h4$+=F4N;h4$+=f9Cm$.t_T;var N0Y=c0r;N0Y+=A1Z;var f6X=f9Cm$.t_T;f6X+=f9Cm$.l60;f6X+=f9Cm$.l60;f6X+=A5h;var H8v=F7I;H8v+=f9Cm$.l60;H8v+=f9Cm$[23424];H8v+=f9Cm$.l60;var p1X=p3g;p1X+=I7r;p1X+=n3_;p1X+=f9Cm$.J4L;this[p1X](c4K,[B3c,submitParams,action,xhr]);I8Z.j$H();this[H8v](this[B5C][f6X][N0Y]);this[I0J](h4R);if(errorCallback){var Y94=q9Z;Y94+=M_7;Y94+=Q6Q;errorCallback[Y94](this,xhr,err,thrown);}this[G29]([i49,h4$],[xhr,err,thrown,submitParams]);}function _tidy(fn){var w3w="setti";I8Z.m_d();var t8d="lete";var F0g="bServerSide";var A97="essi";var I6n="isplay";var k84="Comp";var n3Z=q1i;n3Z+=B5f;n3Z+=Q6Q;n3Z+=f9Cm$.t_T;var X9Y=f9Cm$[555616];X9Y+=I6n;var N6f=h7u;N6f+=A97;N6f+=L47;var T_k=z5S;T_k+=B1d;T_k+=K18;var O5M=K0U;O5M+=f9Cm$.e08;O5M+=V5s;var t4E=f9Cm$[228782];t4E+=f9Cm$[481343];var r$F=f9Cm$.J4L;r$F+=f9Cm$.e08;r$F+=V5s;var _this=this;var dt=this[f9Cm$.J3V][r$F]?new $[t4E][O5M][T_k](this[f9Cm$.J3V][z6u]):B3c;var ssp=h4R;if(dt){var m5T=w3w;m5T+=h7Q;ssp=dt[m5T]()[C37][u8X][F0g];}if(this[f9Cm$.J3V][N6f]){var f19=C9K;f19+=g8Y;f19+=k84;f19+=t8d;this[X9O](f19,function(){if(ssp){var Z9I=f9Cm$[555616];Z9I+=q7Q;Z9I+=H3b;dt[X9O](Z9I,fn);}else {setTimeout(function(){I8Z.m_d();fn();},N3a);}});return X17;}else if(this[X9Y]() === O8g || this[G7z]() === n3Z){this[X9O](K0N,function(){I8Z.m_d();if(!_this[f9Cm$.J3V][T0W]){setTimeout(function(){if(_this[f9Cm$.J3V]){fn();}},N3a);}else {var i0m=f9Cm$[23424];i0m+=f9Cm$[481343];i0m+=f9Cm$.t_T;_this[i0m](L5$,function(e,json){var e$d='draw';I8Z.m_d();if(ssp && json){dt[X9O](e$d,fn);}else {setTimeout(function(){if(_this[f9Cm$.J3V]){fn();}},N3a);}});}})[h1_]();return X17;}return h4R;}function _weakInArray(name,arr){var O6g=Q6Q;O6g+=d3h;I8Z.j$H();for(var i=C37,ien=arr[O6g];i < ien;i++){if(name == arr[i]){return i;}}return -Y5Y;}var fieldType={create:function(){},disable:function(){},enable:function(){},get:function(){},set:function(){}};var DataTable$3=$[M5D][l__];function _buttonText(conf,textIn){var z8H="e...";var K3F="ad b";var D4r="iv.uplo";var x8S="uploadText";var K_n="Choo";var I2T="se fi";var j6l=s$2;j6l+=f9Cm$.J4L;j6l+=J5n;var F1m=f9Cm$[555616];I8Z.j$H();F1m+=D4r;F1m+=K3F;F1m+=D07;var y_o=j7g;y_o+=p3a;y_o+=f9Cm$.Y87;y_o+=f9Cm$.J4L;if(textIn === B3c || textIn === undefined){var j9E=K_n;j9E+=I2T;j9E+=Q6Q;j9E+=z8H;textIn=conf[x8S] || j9E;}conf[y_o][I$n](F1m)[j6l](textIn);}function _commonUpload(editor,conf,dropCallback,multiple){var d_D='<div class="rendered"></div>';var j_x="input[typ";var Q2V="l upload limi";var Q3$="type=\"file\" ";var Y83="buttonInternal";var R$k="tHide\"";var X$R='<div class="editor_upload">';var n3V="class=\"row";var T5_="s=\"cell limitHide\">";var n03="<div class=\"row second";var q0l='<button class="';var C8E='div.drop span';var L6W="on>";var u4v='input[type=file]';var J0w="<div class=\"cel";var C_h='div.clearValue button';var t5R="<div clas";var e0z='dragleave dragexit';var W$V="\"></but";var w1r="ileReade";var k2H='Drag and drop a file here to upload';var V0X="Cla";var P$O="nder";var K1u='<div class="eu_table">';var p16="nput[type=file]";var Z1G="put ";var U7z="dragDropText";var r6d='<div class="cell clearValue">';var Z18='id';var O5m='div.drop';var n5q='></input>';var k1x='dragover';var Q$0="e=fi";var i68='<div class="drop"><span></span></div>';var A6y="noD";var h0V="dragDrop";var e8_='multiple';var e52='"></button>';var x$z=" class=\"cell\">";var q23=K18;q23+=p16;var O4V=f9Cm$[228782];O4V+=K18;O4V+=T52;var K1t=l_m;K1t+=e3G;var B8f=f9Cm$[23424];B8f+=f9Cm$[481343];var R3e=w29;R3e+=f9Cm$[555616];var T6J=y7j;T6J+=w1r;T6J+=f9Cm$.l60;var J1l=p3g;J1l+=f9Cm$[481343];J1l+=j$2;J1l+=V6p;var R3L=k8A;R3L+=R$r;R3L+=h8B;R3L+=i$Z;var F0a=s36;F0a+=x$z;var t9e=Z4$;t9e+=h8B;t9e+=i$Z;var m8s=t5R;m8s+=T5_;var d51=n03;d51+=u_6;var S97=k8A;S97+=T5B;var O3D=W$V;O3D+=f9Cm$.J4L;O3D+=L6W;var v_K=Y61;v_K+=f9Cm$[481343];v_K+=Z1G;v_K+=Q3$;var k62=J0w;k62+=Q2V;k62+=R$k;k62+=i$Z;var q85=d9M;q85+=n3V;q85+=u_6;var Y8T=g2v;Y8T+=b1Z;var u0d=f07;u0d+=f9Cm$.J3V;u0d+=Y1m;u0d+=f9Cm$.J3V;if(multiple === void C37){multiple=h4R;}var btnClass=editor[u0d][Y8T][Y83];var container=$(X$R + K1u + q85 + k62 + q0l + btnClass + e52 + v_K + (multiple?e8_:j_l) + n5q + Q4m + r6d + q0l + btnClass + O3D + Q4m + S97 + d51 + m8s + i68 + t9e + F0a + d_D + Q4m + Q4m + R3L + Q4m);conf[P9Y]=container;conf[J1l]=X17;if(conf[U7A]){var E4y=f9Cm$.J3V;E4y+=Y8G;E4y+=d5P;var Q5u=j_x;Q5u+=Q$0;Q5u+=k_d;Q5u+=M3E;var A7K=f9Cm$[228782];A7K+=K18;A7K+=f9Cm$[481343];A7K+=f9Cm$[555616];container[A7K](Q5u)[S9h](Z18,Editor[E4y](conf[U7A]));}if(conf[S9h]){var o86=f9Cm$.e08;o86+=f9Cm$.J4L;o86+=f9Cm$.J4L;o86+=f9Cm$.l60;container[I$n](u4v)[o86](conf[S9h]);}_buttonText(conf);if(window[T6J] && conf[h0V] !== h4R){var R_2=f9Cm$[23424];R_2+=f9Cm$[481343];var P2Q=f9Cm$[23424];P2Q+=B1d;P2Q+=f9Cm$.t_T;P2Q+=f9Cm$[481343];var Q3d=f9Cm$[23424];Q3d+=f9Cm$[481343];var j2n=f9Cm$[23424];j2n+=f9Cm$[481343];var t09=f9Cm$[555616];t09+=f9Cm$.l60;t09+=f9Cm$[23424];t09+=B1d;var I2Y=f9Cm$[23424];I2Y+=f9Cm$[481343];var U5N=f9Cm$[228782];U5N+=K18;U5N+=f9Cm$[481343];U5N+=f9Cm$[555616];var f1P=f9Cm$.J4L;f1P+=f9Cm$.t_T;f1P+=G1H;f1P+=f9Cm$.J4L;container[I$n](C8E)[f1P](conf[U7z] || k2H);var dragDrop_1=container[U5N](O5m);dragDrop_1[I2Y](t09,function(e){var a5x="original";var W6s="sfe";var L0d="dataTran";if(conf[i43]){var l$b=f9Cm$[23424];l$b+=a6R;l$b+=f9Cm$.l60;var K1i=F3o;K1i+=Q6Q;K1i+=B12;var H_a=L0d;H_a+=W6s;H_a+=f9Cm$.l60;var R11=a5x;R11+=A05;var T_9=f9Cm$.Y87;T_9+=B1d;T_9+=v8S;Editor[T_9](editor,conf,e[R11][H_a][K1i],_buttonText,dropCallback);dragDrop_1[B_c](l$b);}return h4R;})[j2n](e0z,function(e){I8Z.m_d();var v4H="eCla";if(conf[i43]){var r78=o3Z;r78+=f9Cm$.l60;var J1N=O0W;J1N+=l$I;J1N+=v4H;J1N+=i3q;dragDrop_1[J1N](r78);}return h4R;})[h8a](k1x,function(e){if(conf[i43]){var z9S=f9Cm$[23424];z9S+=I7r;z9S+=F7I;dragDrop_1[k1$](z9S);}return h4R;});editor[Q3d](P2Q,function(){var W9h='dragover.DTE_Upload drop.DTE_Upload';I8Z.j$H();var B90=f9Cm$[23424];B90+=f9Cm$[481343];$(p9T)[B90](W9h,function(e){return h4R;});})[R_2](K0N,function(){var h$9="dragover.DTE_U";var d$A="ad drop.DTE_Upload";var v8U=h$9;v8U+=r3r;v8U+=d$A;var W0E=f9Cm$[23424];W0E+=f9Cm$[228782];W0E+=f9Cm$[228782];$(p9T)[W0E](v8U);});}else {var E_d=B2B;E_d+=O4Y;E_d+=P$O;E_d+=w3$;var O3O=C$g;O3O+=p7O;O3O+=f9Cm$[481343];O3O+=f9Cm$[555616];var e$o=A6y;e$o+=f9Cm$.l60;e$o+=f9Cm$[23424];e$o+=B1d;var h$_=O1b;h$_+=f9Cm$[555616];h$_+=V0X;h$_+=i3q;container[h$_](e$o);container[O3O](container[I$n](E_d));}container[R3e](C_h)[B8f](K1t,function(e){var G8i="Defau";var P98=L5f;P98+=G8i;P98+=O6R;I8Z.j$H();e[P98]();if(conf[i43]){var S8f=q9Z;S8f+=f9Cm$.e08;S8f+=Q6Q;S8f+=Q6Q;upload[D0u][S8f](editor,conf,j_l);}});container[O4V](q23)[h8a](N$r,function(){I8Z.j$H();var P3r=F3o;P3r+=G0R;Editor[h5n](editor,conf,this[P3r],_buttonText,function(ids,error){I8Z.m_d();var N8F="input[ty";var a1M="pe=file]";var g4W=P1N;g4W+=f9Cm$.Y87;g4W+=f9Cm$.t_T;var W_B=N8F;W_B+=a1M;var k8c=F3o;k8c+=T52;if(!error){var F_l=q9Z;F_l+=f9Cm$.e08;F_l+=Q6Q;F_l+=Q6Q;dropCallback[F_l](editor,ids);}container[k8c](W_B)[C37][g4W]=j_l;});});return container;}function _triggerChange(input){setTimeout(function(){var d7z="ger";var M2T=Y7m;M2T+=K18;M2T+=h7H;M2T+=d7z;input[M2T](z_j,{editor:X17,editorSet:X17});;},C37);}var baseFieldType=$[E5x](X17,{},fieldType,{canReturnSubmit:function(conf,node){return X17;},disable:function(conf){var H6w=z7h;I8Z.m_d();H6w+=H9V;conf[H6w][j8$](H3i,X17);},enable:function(conf){var F5n="sabled";var Z0e=h44;Z0e+=F5n;var R7D=u6g;R7D+=f9Cm$[23424];R7D+=B1d;var N2c=z7h;N2c+=z5l;N2c+=g7i;conf[N2c][R7D](Z0e,h4R);},get:function(conf){var E5s=T6L;E5s+=B1d;E5s+=g7i;return conf[E5s][P1N]();},set:function(conf,val){var i$A=j7g;i$A+=B96;i$A+=g$w;i$A+=f9Cm$.J4L;I8Z.m_d();var k3j=I7r;k3j+=f9Cm$.e08;k3j+=Q6Q;conf[P9Y][k3j](val);_triggerChange(conf[i$A]);}});var hidden={create:function(conf){var C0X=I7r;C0X+=M_7;C0X+=f9Cm$.Y87;I8Z.j$H();C0X+=f9Cm$.t_T;var U40=n3g;U40+=Q6Q;conf[P9Y]=$(c7g);conf[U40]=conf[C0X];return B3c;},get:function(conf){var i7x=N9S;i7x+=M_7;return conf[i7x];},set:function(conf,val){var q2m=I7r;q2m+=f9Cm$.e08;q2m+=Q6Q;var c0l=z7h;I8Z.j$H();c0l+=i_5;c0l+=f9Cm$.J4L;var oldVal=conf[w2C];conf[w2C]=val;conf[c0l][q2m](val);if(oldVal !== val){_triggerChange(conf[P9Y]);}}};var readonly=$[d7H](X17,{},baseFieldType,{create:function(conf){var g3w="adonl";var m1l="fe";var U62=j7g;U62+=p3a;U62+=g7i;var m6_=x22;m6_+=f9Cm$.l60;I8Z.j$H();var B5i=f9Cm$.l60;B5i+=f9Cm$.t_T;B5i+=g3w;B5i+=g5f;var I8h=K18;I8h+=f9Cm$[555616];var r57=D4Z;r57+=m1l;r57+=d5P;var Q1G=X2q;Q1G+=f9Cm$.J4L;Q1G+=f9Cm$.t_T;Q1G+=T52;var x6B=x22;x6B+=f9Cm$.l60;var c6B=j7g;c6B+=M_q;c6B+=f9Cm$.J4L;conf[c6B]=$(c7g)[x6B]($[Q1G]({id:Editor[r57](conf[I8h]),readonly:B5i,type:f4B},conf[m6_] || ({})));return conf[U62][C37];}});var text=$[d7H](X17,{},baseFieldType,{create:function(conf){var f4o=f9Cm$.e08;f4o+=f9Cm$.J4L;f4o+=f9Cm$.J4L;f4o+=f9Cm$.l60;var Y$d=K18;Y$d+=f9Cm$[555616];var z3L=f9Cm$.J3V;z3L+=Y8G;z3L+=d5P;var n9M=f9Cm$.e08;n9M+=f9Cm$.J4L;n9M+=Y7m;var o$5=j7g;I8Z.j$H();o$5+=B96;o$5+=g$w;o$5+=f9Cm$.J4L;conf[o$5]=$(c7g)[n9M]($[d7H]({id:Editor[z3L](conf[Y$d]),type:f4B},conf[f4o] || ({})));return conf[P9Y][C37];}});var password=$[d7H](X17,{},baseFieldType,{create:function(conf){var b4E="put/";var M3n=j7g;M3n+=K18;M3n+=H9V;var U8d=D6_;U8d+=i3q;U8d+=H3b;U8d+=G5D;var P2n=f9Cm$.t_T;P2n+=G1H;P2n+=f9Cm$.J4L;P2n+=b9X;var K77=f9Cm$.e08;K77+=f9Cm$.J4L;K77+=f9Cm$.J4L;K77+=f9Cm$.l60;var c3G=k8A;c3G+=B96;c3G+=b4E;c3G+=i$Z;conf[P9Y]=$(c3G)[K77]($[P2n]({id:Editor[L2y](conf[U7A]),type:U8d},conf[S9h] || ({})));return conf[M3n][C37];}});var textarea=$[d7H](X17,{},baseFieldType,{canReturnSubmit:function(conf,node){I8Z.j$H();return h4R;},create:function(conf){var R7v='<textarea></textarea>';var y_6=f9Cm$.e08;I8Z.m_d();y_6+=f9Cm$.J4L;y_6+=f9Cm$.J4L;y_6+=f9Cm$.l60;var r3P=f9Cm$.J3V;r3P+=Y8G;r3P+=d5P;var f1D=x22;f1D+=f9Cm$.l60;conf[P9Y]=$(R7v)[f1D]($[d7H]({id:Editor[r3P](conf[U7A])},conf[y_6] || ({})));return conf[P9Y][C37];}});var select=$[d7H](X17,{},baseFieldType,{_addOptions:function(conf,opts,append){var m3H="eholderValue";I8Z.m_d();var L_Z="tionsPair";var N3u="erDisabl";var c7V="ceholder";var g9k="hidden";var d5$="placeholderValue";var Q96="placehold";var v9Q="pai";var j5g="placeholderDisabled";var f98=f9Cm$[23424];f98+=C8a;var n_i=z7h;n_i+=H9V;if(append === void C37){append=h4R;}var elOpts=conf[n_i][C37][f98];var countOffset=C37;if(!append){var n5W=b5Q;n5W+=c7V;elOpts[B97]=C37;if(conf[n5W] !== undefined){var E7v=Q96;E7v+=N3u;E7v+=w3$;var l0e=Q96;l0e+=F7I;var T54=b5Q;T54+=q9Z;T54+=m3H;var placeholderValue=conf[d5$] !== undefined?conf[T54]:j_l;countOffset+=Y5Y;elOpts[C37]=new Option(conf[l0e],placeholderValue);var disabled=conf[E7v] !== undefined?conf[j5g]:X17;elOpts[C37][g9k]=disabled;elOpts[C37][n0f]=disabled;elOpts[C37][B6n]=placeholderValue;}}else {countOffset=elOpts[B97];}if(opts){var l3r=f9Cm$[23424];l3r+=B1d;l3r+=L_Z;var n6Y=v9Q;n6Y+=h_u;Editor[n6Y](opts,conf[l3r],function(val,label,i,attr){var option=new Option(label,val);option[B6n]=val;if(attr){var d2Z=f9Cm$.e08;d2Z+=f9Cm$.J4L;d2Z+=f9Cm$.J4L;d2Z+=f9Cm$.l60;$(option)[d2Z](attr);}elOpts[i + countOffset]=option;});}},create:function(conf){var y3p="feI";var r6M="ipOpts";var o18="ct>";var W55="<select></sele";var A$N=T6L;A$N+=B1d;A$N+=f9Cm$.Y87;A$N+=f9Cm$.J4L;var y2F=W6q;y2F+=n59;y2F+=f9c;var J8d=f9Cm$.J3V;J8d+=f9Cm$.e08;J8d+=y3p;J8d+=f9Cm$[555616];var a8O=f9Cm$.t_T;a8O+=G1H;a8O+=R6y;a8O+=f9Cm$[555616];var H9r=W55;H9r+=o18;var r8Q=T6L;r8Q+=B1d;r8Q+=g7i;conf[r8Q]=$(H9r)[S9h]($[a8O]({id:Editor[J8d](conf[U7A]),multiple:conf[Q$N] === X17},conf[S9h] || ({})))[h8a](I_o,function(e,d){var J72="_las";I8Z.m_d();var t5e="Set";var G5l=w3$;G5l+=g8Y;G5l+=A5h;if(!d || !d[G5l]){var h$C=J72;h$C+=f9Cm$.J4L;h$C+=t5e;conf[h$C]=select[W7y](conf);}});select[Z6J](conf,conf[y2F] || conf[r6M]);return conf[A$N][C37];},destroy:function(conf){var K60=f9Cm$[23424];K60+=f9Cm$[228782];I8Z.m_d();K60+=f9Cm$[228782];conf[P9Y][K60](I_o);},get:function(conf){var S0j="elected";var i7c="eparato";var E5U="option:s";var u9g="rator";var j80="sepa";var v6J=k_d;v6J+=L47;v6J+=f9Cm$.J4L;v6J+=s$2;var H_K=C2U;H_K+=B4Y;H_K+=B1d;H_K+=k_d;var Z_f=p9s;Z_f+=B1d;var k06=E5U;k06+=S0j;I8Z.j$H();var val=conf[P9Y][I$n](k06)[Z_f](function(){I8Z.j$H();return this[B6n];})[f8d]();if(conf[H_K]){var Y1r=f9Cm$.J3V;Y1r+=i7c;Y1r+=f9Cm$.l60;var N0r=j80;N0r+=u9g;return conf[N0r]?val[R5X](conf[Y1r]):val;}return val[v6J]?val[C37]:B3c;},set:function(conf,val,localUpdate){var J6w="placeholder";var m1Q='option';var k7U="isArra";var Z8b="tSe";var G32="selected";var c5i="epar";var k9$="_l";var C4C="ato";var G7p=k_d;G7p+=f9Cm$[481343];G7p+=h5h;var a5t=f9Cm$.t_T;a5t+=f9Cm$.e08;a5t+=j1L;var W0F=F3o;W0F+=T52;var p4g=f9Cm$[228782];p4g+=B96;p4g+=f9Cm$[555616];var U_k=k7U;U_k+=g5f;if(!localUpdate){var o_F=k9$;o_F+=R77;o_F+=Z8b;o_F+=f9Cm$.J4L;conf[o_F]=val;}if(conf[Q$N] && conf[v9Y] && !Array[U_k](val)){var p_v=f9Cm$.J3V;p_v+=c5i;p_v+=C4C;p_v+=f9Cm$.l60;var u83=V1c;u83+=r0a;var v0T=f9Cm$.J3V;v0T+=Y7m;v0T+=K18;v0T+=L47;val=typeof val === v0T?val[u83](conf[p_v]):[];}else if(!Array[d2_](val)){val=[val];}var i;var len=val[B97];var found;var allFound=h4R;var options=conf[P9Y][p4g](m1Q);conf[P9Y][W0F](m1Q)[a5t](function(){I8Z.m_d();found=h4R;for(i=C37;i < len;i++){if(this[B6n] == val[i]){found=X17;allFound=X17;break;}}this[G32]=found;});if(conf[J6w] && !allFound && !conf[Q$N] && options[G7p]){options[C37][G32]=X17;}if(!localUpdate){var g2o=j7g;g2o+=B96;g2o+=B1d;g2o+=g7i;_triggerChange(conf[g2o]);}I8Z.m_d();return allFound;},update:function(conf,options,append){var Q8y="tSet";var L8x=j7g;L8x+=Q6Q;L8x+=R77;I8Z.m_d();L8x+=Q8y;select[Z6J](conf,options,append);var lastSet=conf[L8x];if(lastSet !== undefined){select[D0u](conf,lastSet,X17);}_triggerChange(conf[P9Y]);}});var checkbox=$[d7H](X17,{},baseFieldType,{_addOptions:function(conf,opts,append){var f0_="onsPa";if(append === void C37){append=h4R;}var jqInput=conf[P9Y];var offset=C37;if(!append){jqInput[Z2L]();}else {var Y4B=Q6Q;Y4B+=f9Cm$.t_T;Y4B+=L47;Y4B+=K5v;var L2V=p3a;L2V+=g7i;offset=$(L2V,jqInput)[Y4B];}if(opts){var T8E=O11;T8E+=f0_;T8E+=w7h;Editor[O8s](opts,conf[T8E],function(val,label,i,attr){var V9C="<la";var R$i="</l";var Q2X=" id=\"";var h4L=":last";var r8N="abel>";var D8d="<in";var B6d='" type="checkbox" />';var i$j="bel for=\"";var S$x=p9D;S$x+=n1E;S$x+=Q6Q;var t1G=K18;t1G+=z5l;t1G+=B_I;I8Z.j$H();t1G+=D1I;var C0$=R$i;C0$+=r8N;var T6w=y9e;T6w+=i$Z;var b_O=v0N;b_O+=f9Cm$.t_T;b_O+=d5P;var j7e=V9C;j7e+=i$j;var K$x=v0N;K$x+=f9Cm$.t_T;K$x+=h3G;K$x+=f9Cm$[555616];var f2s=D8d;f2s+=g$w;f2s+=f9Cm$.J4L;f2s+=Q2X;var i96=q7L;i96+=K18;i96+=T5I;var M6Y=C$g;M6Y+=B1d;M6Y+=b9X;jqInput[M6Y](i96 + f2s + Editor[K$x](conf[U7A]) + X_H + (i + offset) + B6d + j7e + Editor[b_O](conf[U7A]) + X_H + (i + offset) + T6w + label + C0$ + Q4m);$(t1G,jqInput)[S9h](T1q,val)[C37][S$x]=val;if(attr){var C1E=K18;C1E+=H9V;C1E+=h4L;$(C1E,jqInput)[S9h](attr);}});}},create:function(conf){var K6G="ipO";var X4D="></di";var a0k=T6L;a0k+=p_t;var u8U=K6G;u8U+=g$5;var g6j=s36;g6j+=X4D;g6j+=T5I;conf[P9Y]=$(g6j);checkbox[Z6J](conf,conf[O$K] || conf[u8U]);return conf[a0k][C37];},disable:function(conf){var d1K=f9Cm$[555616];d1K+=K_b;I8Z.m_d();d1K+=Z_K;var x2l=K18;x2l+=f9Cm$[481343];x2l+=p_t;var N6B=T6L;N6B+=B1d;N6B+=g7i;conf[N6B][I$n](x2l)[j8$](d1K,X17);},enable:function(conf){var o64=f9Cm$[228782];o64+=K18;o64+=f9Cm$[481343];o64+=f9Cm$[555616];I8Z.m_d();conf[P9Y][o64](N$r)[j8$](H3i,h4R);},get:function(conf){var U$G="arat";var U3o="parato";var x9r="sep";var G7G="nselectedValue";var h_F=x9r;h_F+=U$G;h_F+=f9Cm$[23424];h_F+=f9Cm$.l60;var w6t=Y1m;w6t+=U3o;w6t+=f9Cm$.l60;var W2k=f9Cm$.Y87;W2k+=G7G;var I0T=Q6Q;I0T+=q8V;I0T+=f9Cm$.J4L;I0T+=s$2;var A23=F3o;A23+=f9Cm$[481343];A23+=f9Cm$[555616];I8Z.m_d();var U7E=j7g;U7E+=K18;U7E+=z5l;U7E+=g7i;var out=[];var selected=conf[U7E][A23](C10);if(selected[I0T]){var b7c=p3M;b7c+=s$2;selected[b7c](function(){var q3w=B1d;q3w+=f9Cm$.Y87;q3w+=f9Cm$.J3V;q3w+=s$2;out[q3w](this[B6n]);});}else if(conf[W2k] !== undefined){out[Z_J](conf[W99]);}return conf[w6t] === undefined || conf[v9Y] === B3c?out:out[R5X](conf[h_F]);},set:function(conf,val){var m_s="rin";var Y6b='|';var I6T="isAr";var M9x="rray";var S7T=z5H;S7T+=M9x;var m0k=f9Cm$.J3V;m0k+=f9Cm$.J4L;m0k+=m_s;m0k+=h7H;var Y8w=I6T;Y8w+=q7Q;Y8w+=g5f;var r2I=T6L;r2I+=B1d;r2I+=f9Cm$.Y87;r2I+=f9Cm$.J4L;var jqInputs=conf[r2I][I$n](N$r);if(!Array[Y8w](val) && typeof val === m0k){var O_U=f9Cm$.J3V;O_U+=B1d;O_U+=r0a;val=val[O_U](conf[v9Y] || Y6b);}else if(!Array[S7T](val)){val=[val];}var i;var len=val[B97];var found;jqInputs[a8N](function(){var P$W=R_f;P$W+=e6L;found=h4R;for(i=C37;i < len;i++){var s3n=j7g;s3n+=C6T;s3n+=n1E;s3n+=Q6Q;if(this[s3n] == val[i]){found=X17;break;}}this[P$W]=found;});I8Z.j$H();_triggerChange(jqInputs);},update:function(conf,options,append){var O1e="Opt";var D9v="_ad";var x2R=f9Cm$.J3V;x2R+=f9Cm$.t_T;x2R+=f9Cm$.J4L;var u2M=D9v;u2M+=f9Cm$[555616];u2M+=O1e;u2M+=g4D;var p7j=h7H;p7j+=f9Cm$.t_T;p7j+=f9Cm$.J4L;var currVal=checkbox[p7j](conf);checkbox[u2M](conf,options,append);checkbox[x2R](conf,currVal);}});var radio=$[d7H](X17,{},baseFieldType,{_addOptions:function(conf,opts,append){var Z0C="emp";if(append === void C37){append=h4R;}var jqInput=conf[P9Y];var offset=C37;I8Z.m_d();if(!append){var x4y=Z0C;x4y+=q_4;jqInput[x4y]();}else {var t5_=k_d;t5_+=g7N;t5_+=s$2;offset=$(N$r,jqInput)[t5_];}if(opts){Editor[O8s](opts,conf[d_v],function(val,label,i,attr){var M1A='" />';var V0J="ame=";var a2y='input:last';var O9f='<div>';var O$j=" type=\"radio\" n";var p5W='<label for="';var n$Z='<input id="';var i$p="itor_";var D$U=p8d;D$U+=i$p;D$U+=P1N;var F2x=f9Cm$.e08;F2x+=f9Cm$.J4L;F2x+=Y7m;var j9J=N1z;j9J+=f9Cm$.t_T;var R6E=y9e;R6E+=O$j;R6E+=V0J;R6E+=y9e;var t2I=v0N;t2I+=f9Cm$.t_T;t2I+=h3G;t2I+=f9Cm$[555616];var Q04=f9Cm$.e08;Q04+=Y8i;I8Z.j$H();Q04+=b9X;jqInput[Q04](O9f + n$Z + Editor[t2I](conf[U7A]) + X_H + (i + offset) + R6E + conf[j9J] + M1A + p5W + Editor[L2y](conf[U7A]) + X_H + (i + offset) + x$l + label + t7l + Q4m);$(a2y,jqInput)[F2x](T1q,val)[C37][D$U]=val;if(attr){var X4d=B96;X4d+=B1d;X4d+=B_I;X4d+=D1I;$(X4d,jqInput)[S9h](attr);}});}},create:function(conf){var Q0m="ipOpt";var D8l="ope";var H9l="/>";var x3y="div ";var a0R=D8l;a0R+=f9Cm$[481343];var l9P=f9Cm$[23424];l9P+=f9Cm$[481343];var n6V=Q0m;n6V+=f9Cm$.J3V;var Y$C=k8A;Y$C+=x3y;Y$C+=H9l;conf[P9Y]=$(Y$C);radio[Z6J](conf,conf[O$K] || conf[n6V]);this[l9P](a0R,function(){var r$3=f9Cm$.t_T;r$3+=f9Cm$.e08;r$3+=q9Z;r$3+=s$2;var C8Z=B96;C8Z+=g$w;C8Z+=f9Cm$.J4L;var d7L=F3o;I8Z.m_d();d7L+=f9Cm$[481343];d7L+=f9Cm$[555616];conf[P9Y][d7L](C8Z)[r$3](function(){var o9D="ecked";var H7S="Checked";var D9G="_pr";var Q2E=D9G;Q2E+=f9Cm$.t_T;Q2E+=H7S;if(this[Q2E]){var b6a=j1L;b6a+=o9D;this[b6a]=X17;}});});I8Z.j$H();return conf[P9Y][C37];},disable:function(conf){var N8p=h44;N8p+=f9Cm$.J3V;N8p+=Z_K;var u2n=B1d;u2n+=f9Cm$.l60;u2n+=J_q;var g0K=p3a;g0K+=g7i;var d9n=F3o;d9n+=f9Cm$[481343];d9n+=f9Cm$[555616];conf[P9Y][d9n](g0K)[u2n](N8p,X17);},enable:function(conf){var E6L="rop";var Z7k=B1d;Z7k+=E6L;I8Z.m_d();var g_P=B96;g_P+=B1d;g_P+=g7i;var F2K=f9Cm$[228782];F2K+=K18;F2K+=f9Cm$[481343];F2K+=f9Cm$[555616];var Z$K=T6L;Z$K+=B1d;Z$K+=g7i;conf[Z$K][F2K](g_P)[Z7k](H3i,h4R);},get:function(conf){var J6M="tedVa";var h0A="unselec";var I3E=h0A;I8Z.m_d();I3E+=J6M;I3E+=c5p;I3E+=f9Cm$.t_T;var q8I=Q6Q;q8I+=q8V;q8I+=f9Cm$.J4L;q8I+=s$2;var B71=f9Cm$[228782];B71+=E$W;var el=conf[P9Y][B71](C10);if(el[q8I]){return el[C37][B6n];}return conf[I3E] !== undefined?conf[W99]:undefined;},set:function(conf,val){var f7c=":check";var y0h=S8e;y0h+=f7c;y0h+=w3$;var F3Q=f9Cm$[228782];F3Q+=K18;F3Q+=f9Cm$[481343];F3Q+=f9Cm$[555616];conf[P9Y][F3Q](N$r)[a8N](function(){var x_h="ked";var x7o="_preChec";var t$A="_preChecked";var c7o="checked";var O0q="or_val";var k5_="preCh";var t9J=j7g;t9J+=w3$;t9J+=g8Y;t9J+=O0q;var L88=x7o;I8Z.j$H();L88+=x_h;this[L88]=h4R;if(this[t9J] == val){var b4c=j7g;b4c+=k5_;b4c+=d6G;b4c+=x_h;var j1n=R_f;j1n+=e6L;this[j1n]=X17;this[b4c]=X17;}else {this[c7o]=h4R;this[t$A]=h4R;}});_triggerChange(conf[P9Y][I$n](y0h));},update:function(conf,options,append){var d8h='[value="';var P4c="eq";var v5J=P1N;v5J+=f9Cm$.Y87;v5J+=f9Cm$.t_T;var v4v=y9e;I8Z.j$H();v4v+=M3E;var J3F=f9Cm$.J3V;J3F+=f9Cm$.t_T;J3F+=f9Cm$.J4L;var h9F=B96;h9F+=B1d;h9F+=g7i;var X9T=z7h;X9T+=z5l;X9T+=f9Cm$.Y87;X9T+=f9Cm$.J4L;var D76=h7H;D76+=f9Cm$.t_T;D76+=f9Cm$.J4L;var currVal=radio[D76](conf);radio[Z6J](conf,options,append);var inputs=conf[X9T][I$n](h9F);radio[J3F](conf,inputs[e3Q](d8h + currVal + v4v)[B97]?currVal:inputs[P4c](C37)[S9h](v5J));}});var datetime=$[I1i](X17,{},baseFieldType,{create:function(conf){var y2q="_inp";var X40="_closeFn";var d17="entLocal";var m0K="eTime";var k7y="ale";var c6u="locale";var u$l="seFn";var r$$="eI";var W2c="_cl";I8Z.m_d();var J7o="momentStrict";var F36="strict";var e6J="datetim";var m7L="mom";var g77="Dat";var y0T='DateTime library is required';var d8e="ntStrict";var F7N="put /";var y0S="Local";var J3M="oment";var i3s="stri";var S$J="yInput";var v4A="displayFormat";var B3U=f9Cm$[23424];B3U+=f9Cm$[481343];var t1O=i3K;t1O+=S$J;var I90=W2c;I90+=f9Cm$[23424];I90+=u$l;var t2V=f9Cm$[23424];t2V+=B1d;t2V+=f9Cm$.E9M;var H0x=e6J;H0x+=f9Cm$.t_T;var W9z=i$c;W9z+=D6p;W9z+=f9Cm$.e08;W9z+=f9Cm$.J4L;var f3R=H_L;f3R+=b9X;var s1x=l93;s1x+=o2e;s1x+=u08;s1x+=K6S;var W7H=W6q;W7H+=f9Cm$.J3V;var y8u=m7L;y8u+=f9Cm$.t_T;y8u+=d8e;var j7v=F7s;j7v+=q9Z;j7v+=k7y;var e7p=m7L;e7p+=d17;e7p+=f9Cm$.t_T;var d9Z=g77;d9Z+=m0K;var W9Q=U58;W9Q+=M27;var x35=K18;x35+=f9Cm$[555616];var a7i=D4Z;a7i+=f9Cm$[228782];a7i+=r$$;a7i+=f9Cm$[555616];var S4L=Y61;S4L+=f9Cm$[481343];S4L+=F7N;S4L+=i$Z;var b5J=y2q;b5J+=g7i;conf[b5J]=$(S4L)[S9h]($[d7H](X17,{id:Editor[a7i](conf[x35]),type:W9Q},conf[S9h]));if(!DataTable$3[d9Z]){var p0G=y3N;p0G+=A5h;Editor[p0G](y0T,T5p);}if(conf[e7p] && !conf[w3u][j7v]){var T4X=D6p;T4X+=J3M;T4X+=y0S;T4X+=f9Cm$.t_T;conf[w3u][c6u]=conf[T4X];}if(conf[y8u] && !conf[W7H][F36]){var j1E=i3s;j1E+=q9Z;j1E+=f9Cm$.J4L;var K3h=J_q;K3h+=f9Cm$.E9M;conf[K3h][j1E]=conf[J7o];}conf[m3u]=new DataTable$3[s1x](conf[P9Y],$[f3R]({format:conf[v4A] || conf[W9z],i18n:this[B5C][H0x]},conf[t2V]));conf[I90]=function(){var G$L="ker";var s$K=j7g;s$K+=B1d;s$K+=B1B;s$K+=G$L;conf[s$K][J7j]();};if(conf[t1O] === h4R){var X7F=f9Cm$[23424];X7F+=f9Cm$[481343];var u51=j7g;u51+=B96;u51+=p_t;conf[u51][X7F](w_k,function(e){I8Z.m_d();e[G8l]();});}this[B3U](K0N,conf[X40]);return conf[P9Y][C37];},destroy:function(conf){var X6p="icke";var a5o="keydo";var A33="eF";var D2V=q_Z;D2V+=X6p;D2V+=f9Cm$.l60;var K_Z=a5o;K_Z+=H3b;K_Z+=f9Cm$[481343];var Q60=f9Cm$[23424];Q60+=f9Cm$[228782];Q60+=f9Cm$[228782];var M7f=h7m;M7f+=f9Cm$.J3V;M7f+=A33;M7f+=f9Cm$[481343];var m2j=f9Cm$[23424];m2j+=f9Cm$[228782];m2j+=f9Cm$[228782];I8Z.m_d();this[m2j](K0N,conf[M7f]);conf[P9Y][Q60](K_Z);conf[D2V][J8j]();},errorMessage:function(conf,msg){var k0S="errorMsg";I8Z.m_d();conf[m3u][k0S](msg);},get:function(conf){I8Z.j$H();var f4d="valFormat";var I$t="reFor";var s1r="_picke";var l63=z7h;l63+=f9Cm$[481343];l63+=B1d;l63+=g7i;var x62=H3b;x62+=K18;x62+=I$t;x62+=Y0o;var o3m=s1r;o3m+=f9Cm$.l60;return conf[G27]?conf[o3m][f4d](conf[x62]):conf[l63][P1N]();},maxDate:function(conf,max){var L0v="_pi";var H8r="cke";var B3M=D6p;B3M+=f9Cm$.e08;B3M+=G1H;var y86=L0v;y86+=H8r;y86+=f9Cm$.l60;conf[y86][B3M](max);},minDate:function(conf,min){var R9h=D6p;R9h+=K18;R9h+=f9Cm$[481343];I8Z.m_d();var Z9b=j7g;Z9b+=h0o;conf[Z9b][R9h](min);},owns:function(conf,node){var C4x=f9Cm$[23424];C4x+=H3b;C4x+=f9Cm$[481343];C4x+=f9Cm$.J3V;var M6h=j7g;M6h+=h0o;return conf[M6h][C4x](node);},set:function(conf,val){var R55="lF";var t9y='--';var B1S="va";var d64="_pick";I8Z.m_d();var V_B=T8i;V_B+=f9Cm$.J4L;var m1e=E$W;m1e+=X2q;m1e+=v0F;var B_8=f9Cm$.J3V;B_8+=J2x;B_8+=L47;if(typeof val === B_8 && val && val[m1e](t9y) !== C37 && conf[G27]){var D4x=B1S;D4x+=R55;D4x+=M55;D4x+=f9Cm$.J4L;conf[m3u][D4x](conf[G27],val);}else {var S2I=I7r;S2I+=M_7;var o3J=d64;o3J+=F7I;conf[o3J][S2I](val);}_triggerChange(conf[V_B]);}});var upload=$[d7H](X17,{},baseFieldType,{canReturnSubmit:function(conf,node){I8Z.m_d();return h4R;},create:function(conf){var editor=this;var container=_commonUpload(editor,conf,function(val){var q3e="even";var I47=f9Cm$[481343];I47+=f9Cm$.e08;I47+=W1R;var t1s=j7g;t1s+=q3e;t1s+=f9Cm$.J4L;var R3s=q9Z;R3s+=f9Cm$.e08;R3s+=Q6Q;R3s+=Q6Q;var K3Z=f9Cm$.J3V;K3Z+=C6V;upload[K3Z][R3s](editor,conf,val[C37]);editor[t1s](v3Q,[conf[I47],val[C37]]);});return container;},disable:function(conf){var s3$="disabl";var k93=s3$;k93+=f9Cm$.t_T;k93+=f9Cm$[555616];var Z2P=p3a;Z2P+=f9Cm$.Y87;Z2P+=f9Cm$.J4L;I8Z.j$H();var H03=f9Cm$[228782];H03+=K18;H03+=f9Cm$[481343];H03+=f9Cm$[555616];conf[P9Y][H03](Z2P)[j8$](k93,X17);conf[i43]=h4R;},enable:function(conf){var U4Y=u6g;U4Y+=J_q;var q2S=K18;q2S+=z5l;q2S+=f9Cm$.Y87;q2S+=f9Cm$.J4L;var T6S=f9Cm$[228782];T6S+=K18;T6S+=f9Cm$[481343];T6S+=f9Cm$[555616];conf[P9Y][T6S](q2S)[U4Y](H3i,h4R);conf[i43]=X17;},get:function(conf){I8Z.m_d();var b01=n3g;b01+=Q6Q;return conf[b01];},set:function(conf,val){var t0K="oFile";var B7T="file";var W4J="o ";var m7Q="div.rend";var G82="button";var C9M="clearText";var H3Z="Tex";var T2Z="div.clearValue ";var W8$='noClear';var h1r="learText";var J0E=j7g;J0E+=I7r;J0E+=f9Cm$.e08;J0E+=Q6Q;var G76=h5n;G76+=I9t;G76+=W9u;var e1H=K18;e1H+=f9Cm$[481343];e1H+=B1d;e1H+=g7i;var t5Y=f9Cm$[228782];I8Z.j$H();t5Y+=E$W;var W6V=q9Z;W6V+=h1r;var r8A=T2Z;r8A+=G82;var R_K=f9Cm$[228782];R_K+=K18;R_K+=T52;var b50=b7s;b50+=Q6Q;b50+=n7x;var J2Y=T8i;J2Y+=f9Cm$.J4L;var T7c=I7r;T7c+=f9Cm$.e08;T7c+=Q6Q;var b89=z7h;b89+=i_5;b89+=f9Cm$.J4L;conf[w2C]=val;conf[b89][T7c](j_l);var container=conf[J2Y];if(conf[b50]){var V$a=j7g;V$a+=I7r;V$a+=f9Cm$.e08;V$a+=Q6Q;var z6D=m7Q;z6D+=F7I;z6D+=w3$;var rendered=container[I$n](z6D);if(conf[V$a]){var x$J=j7g;x$J+=I7r;x$J+=f9Cm$.e08;x$J+=Q6Q;var L4k=h44;L4k+=t6u;var Z2Z=h0z;Z2Z+=Q6Q;rendered[Z2Z](conf[L4k](conf[x$J]));}else {var o1K=Y4L;o1K+=W4J;o1K+=B7T;var u_E=f9Cm$[481343];u_E+=t0K;u_E+=H3Z;u_E+=f9Cm$.J4L;var W$e=f9Cm$.e08;W$e+=B1d;W$e+=p7O;W$e+=T52;rendered[Z2L]()[W$e](j6Z + (conf[u_E] || o1K) + p88);}}var button=container[R_K](r8A);if(val && conf[W6V]){var J6B=h0z;J6B+=Q6Q;button[J6B](conf[C9M]);container[B_c](W8$);}else {container[k1$](W8$);}conf[P9Y][t5Y](e1H)[V39](G76,[conf[J0E]]);}});var uploadMany=$[F3K](X17,{},baseFieldType,{_showHide:function(conf){var a6G=".limitHide";var S7j="_limitLe";var P43="limit";var f$I="limi";var q7G=Q6Q;q7G+=d3h;var Z$M=N9S;Z$M+=f9Cm$.e08;Z$M+=Q6Q;var w1Y=k1J;w1Y+=S_t;var s$P=S7j;s$P+=u3w;var Q6c=j7g;Q6c+=P1N;var u$x=h8B;u$x+=a6G;I8Z.j$H();var F9z=f9Cm$[228782];F9z+=K18;F9z+=T52;var y$Z=f$I;y$Z+=f9Cm$.J4L;if(!conf[y$Z]){return;}conf[H$W][F9z](u$x)[X5f](M_p,conf[Q6c][B97] >= conf[P43]?n50:M20);conf[s$P]=conf[w1Y] - conf[Z$M][q7G];},canReturnSubmit:function(conf,node){return h4R;},create:function(conf){var D8N='multi';var L$5='button.remove';var b1F="ick";var N_i=M_R;N_i+=b1F;var editor=this;var container=_commonUpload(editor,conf,function(val){var O24=f9Cm$[481343];O24+=f9Cm$.e08;O24+=D6p;O24+=f9Cm$.t_T;var s_p=j7g;s_p+=I7r;s_p+=f9Cm$.e08;s_p+=Q6Q;var S_7=q9Z;S_7+=h8a;S_7+=q9Z;S_7+=k6$;var o5D=j7g;o5D+=I7r;o5D+=f9Cm$.e08;o5D+=Q6Q;conf[o5D]=conf[w2C][S_7](val);uploadMany[D0u][B7t](editor,conf,conf[s_p]);editor[G29](v3Q,[conf[O24],conf[w2C]]);},X17);I8Z.j$H();container[k1$](D8N)[h8a](N_i,L$5,function(e){var n$t="opPropagati";var J29="_enabl";var J2C=J29;J2C+=f9Cm$.t_T;J2C+=f9Cm$[555616];var I4E=U5a;I4E+=n$t;I4E+=h8a;e[I4E]();if(conf[J2C]){var Q7A=Y1m;Q7A+=f9Cm$.J4L;var h$Q=j7g;h$Q+=I7r;h$Q+=f9Cm$.e08;h$Q+=Q6Q;var v_a=K18;v_a+=f9Cm$[555616];v_a+=G1H;var idx=$(this)[I$4](v_a);conf[h$Q][O9l](idx,Y5Y);uploadMany[Q7A][B7t](editor,conf,conf[w2C]);}});conf[H$W]=container;return container;},disable:function(conf){var w$o="sabl";var J$U=j7g;J$U+=y9q;J$U+=f9Cm$[555616];var R0F=h44;R0F+=w$o;R0F+=w3$;var b$X=G2J;b$X+=B1d;var F29=K18;F29+=f9Cm$[481343];F29+=p_t;I8Z.j$H();var D5B=w29;D5B+=f9Cm$[555616];conf[P9Y][D5B](F29)[b$X](R0F,X17);conf[J$U]=h4R;},enable:function(conf){var o68="_en";var x90=o68;x90+=j$2;x90+=V6p;var q0Z=u6g;q0Z+=f9Cm$[23424];q0Z+=B1d;var Q2B=K18;Q2B+=z5l;Q2B+=g7i;var M8N=j7g;M8N+=B96;M8N+=B1d;I8Z.j$H();M8N+=g7i;conf[M8N][I$n](Q2B)[q0Z](H3i,h4R);conf[x90]=X17;},get:function(conf){I8Z.j$H();return conf[w2C];},set:function(conf,val){var h4q="leT";var E_R="dered";var S_G="<ul>";var b6c="No ";var R02="Handl";var J90="</ul>";var S0n="il";var B8l="_showHide";var M7X="noF";var y9P="div.r";var V91="uplo";var H1H='Upload collections must have an array as a value';var s3g=V91;s3g+=O1b;s3g+=u5z;s3g+=Z9K;var v_Y=K7O;v_Y+=F7I;v_Y+=R02;v_Y+=F7I;var i9z=f9Cm$[228782];i9z+=B96;i9z+=f9Cm$[555616];var O1Y=T8i;O1Y+=f9Cm$.J4L;var t5n=h44;t5n+=t6u;var W8I=j7g;W8I+=B96;W8I+=g$w;W8I+=f9Cm$.J4L;var O4Q=I7r;O4Q+=f9Cm$.e08;O4Q+=Q6Q;I8Z.m_d();var o9e=j7g;o9e+=I7r;o9e+=f9Cm$.e08;o9e+=Q6Q;if(!val){val=[];}if(!Array[d2_](val)){throw new Error(H1H);}conf[o9e]=val;conf[P9Y][O4Q](j_l);var that=this;var container=conf[W8I];if(conf[t5n]){var S9v=Q6Q;S9v+=d3h;var K$Y=f9Cm$.t_T;K$Y+=D6p;K$Y+=B1d;K$Y+=q_4;var V5Y=y9P;V5Y+=f9Cm$.t_T;V5Y+=f9Cm$[481343];V5Y+=E_R;var M9R=F3o;M9R+=f9Cm$[481343];M9R+=f9Cm$[555616];var rendered=container[M9R](V5Y)[K$Y]();if(val[S9v]){var i4I=p3M;i4I+=s$2;var i_z=D_8;i_z+=f9Cm$[555616];i_z+=X8A;var U0o=S_G;U0o+=J90;var list_1=$(U0o)[i_z](rendered);$[i4I](val,function(i,file){var d_h="i>";var Q7i="orm";var D5y=" remov";var d8w='">&times;</button>';var v1c="e\" data-idx=\"";var h3t=' <button class="';var x0H='</li>';I8Z.m_d();var display=conf[G7z](file,i);if(display !== B3c){var v0V=D5y;v0V+=v1c;var l8s=E2j;l8s+=L_M;var R1V=f9Cm$[228782];R1V+=Q7i;var O3M=k8A;O3M+=Q6Q;O3M+=d_h;list_1[P2$](O3M + display + h3t + that[A2x][R1V][l8s] + v0V + i + d8w + x0H);}});}else {var R2B=b6c;R2B+=f9Cm$[228782];R2B+=S0n;R2B+=B12;var I6R=M7X;I6R+=K18;I6R+=h4q;I6R+=H_L;rendered[P2$](j6Z + (conf[I6R] || R2B) + p88);}}uploadMany[B8l](conf);conf[O1Y][i9z](N$r)[v_Y](s3g,[conf[w2C]]);}});var datatable=$[Z0w](X17,{},baseFieldType,{_addOptions:function(conf,options,append){var o8w=f9Cm$.e08;o8w+=J2K;var y2A=f9Cm$.l60;y2A+=h98;y2A+=f9Cm$.J3V;var H1O=f9Cm$[555616];H1O+=f9Cm$.J4L;if(append === void C37){append=h4R;}var dt=conf[H1O];I8Z.m_d();if(!append){dt[M1r]();}dt[y2A][o8w](options)[u8I]();},_jumpToFirst:function(conf,editor){var y5Y='applied';var C0L="appl";var w7s="page";var U2h="_scrollB";var Z12="exes";var o3C='open';var u3B="div.dataTables";var V89="arents";var M5$="dexO";var M2F="tainer";var d4P="mber";var Q2c="ody";var o6S="ied";var q$c="loo";var V7E=A6T;V7E+=f9Cm$[481343];V7E+=M2F;var L9s=u3B;L9s+=U2h;L9s+=Q2c;var W_N=f9Cm$[555616];W_N+=f9Cm$.l60;W_N+=f9Cm$.e08;W_N+=H3b;var T2l=B1d;T2l+=f9Cm$.e08;T2l+=h7H;T2l+=f9Cm$.t_T;var c__=f9Cm$[481343];c__+=f9Cm$.Y87;c__+=d4P;var h_7=C0L;h_7+=o6S;var c6S=f9Cm$.l60;c6S+=h98;var X8H=f9Cm$[555616];X8H+=f9Cm$.J4L;var dt=conf[X8H];var idx=dt[c6S]({order:h_7,selected:X17})[R27]();var page=C37;if(typeof idx === c__){var P5x=f9Cm$[228782];P5x+=q$c;P5x+=f9Cm$.l60;var d1d=B96;d1d+=M5$;d1d+=f9Cm$[228782];var V5L=B96;V5L+=f9Cm$[555616];V5L+=Z12;var pageLen=dt[w7s][c0T]()[B97];var pos=dt[X9o]({order:y5Y})[V5L]()[d1d](idx);page=pageLen > C37?Math[P5x](pos / pageLen):C37;}dt[T2l](page)[W_N](h4R);var container=$(L9s,dt[z6u]()[V7E]());var scrollTo=function(){var T_x="scrollTo";I8Z.m_d();var Y7D="sition";var P$m="appli";var I7w=f9Cm$[481343];I7w+=x4Y;var D9l=P$m;D9l+=f9Cm$.t_T;D9l+=f9Cm$[555616];var node=dt[P2R]({order:D9l,selected:X17})[I7w]();if(node){var H7M=f9Cm$.J4L;H7M+=J_q;var b4h=B1d;b4h+=f9Cm$[23424];b4h+=Y7D;var height=container[Z$p]();var top_1=$(node)[b4h]()[H7M];if(top_1 > height - N3a){var x0D=T_x;x0D+=B1d;container[x0D](top_1);}}};if(container[B97]){var z9a=f4W;z9a+=s$2;var I$A=r0K;I$A+=f9Cm$[555616];I$A+=g5f;var l1p=B1d;l1p+=V89;if(container[l1p](I$A)[z9a]){scrollTo();}else {var P0o=f9Cm$[23424];P0o+=f9Cm$[481343];P0o+=f9Cm$.t_T;editor[P0o](o3C,function(){scrollTo();});}}},create:function(conf){var R_t="config";var m4F="<tfo";var D0a="<t";var N3X="r>";var V$8="r-select";var U4s="tionsPa";var G9y='Search';var f8D="use";var c$t="ataT";var m5G='<div class="DTE_Field_Type_datatable_info">';var O4k="Btp";var z3k='<table>';var s6q="mitComplete";var S4a='init.dt';var I1Z='single';var b8q="Pair";var o6g="addO";var Z5P="Arra";var B6v='100%';var K$C="ot>";var E7n=W6q;E7n+=K18;E7n+=W4z;var T2j=j7g;T2j+=o6g;T2j+=C8a;var u6x=f9Cm$.t_T;u6x+=f9Cm$[555616];u6x+=g8Y;u6x+=A5h;var o6L=f8D;o6L+=V$8;var w2D=f9Cm$[23424];w2D+=B1d;w2D+=n3_;var o3c=f9Cm$[23424];o3c+=f9Cm$[481343];var n8e=f9Cm$[228782];n8e+=K18;n8e+=O4k;var u09=D8m;u09+=k0d;var C4v=J_q;C4v+=U4s;C4v+=w7h;var B05=f9Cm$.t_T;B05+=M27;B05+=f9Cm$.t_T;B05+=T52;var q7w=l93;q7w+=c$t;q7w+=l27;var y6Z=f9Cm$[23424];y6Z+=f9Cm$[481343];var b7f=g6l;b7f+=f9Cm$.J4L;b7f+=s$2;var Y9s=z6u;Y9s+=s_b;Y9s+=S1i;Y9s+=i3q;var i9r=X0d;i9r+=Q6Q;i9r+=f9Cm$.e08;i9r+=i3q;var r5$=p26;r5$+=b9X;var N5W=k8A;N5W+=f9Cm$[555616];N5W+=K18;N5W+=T5I;var I93=J_q;I93+=Y$F;I93+=b8q;var M2U=Q6Q;M2U+=f9Cm$.e08;M2U+=h2q;M2U+=Q6Q;var c4q=H_L;c4q+=b9X;var _this=this;conf[d_v]=$[c4q]({label:M2U,value:T1q},conf[I93]);var table=$(z3k);var container=$(N5W)[r5$](table);var side=$(m5G);if(conf[J$g]){var h36=D0a;h36+=N3X;var u7j=K_b;u7j+=Z5P;u7j+=g5f;var C_p=m4F;C_p+=K$C;$(C_p)[P2$](Array[u7j](conf[J$g])?$(h36)[P2$]($[X5_](conf[J$g],function(str){var c1p="<th";var w0r=c1p;I8Z.m_d();w0r+=i$Z;return $(w0r)[n8n](str);})):conf[J$g])[A8d](table);}var dt=table[i9r](datatable[Y9s])[b7f](B6v)[y6Z](S4a,function(e,settings){var p8U="iv.d";var K4G="ataTables";var l4m="ataTa";var l6w="nta";var h5t='div.dt-buttons';var I1d="div.d";var R6h="nTabl";var O7d="bles_info";var X2z="_fil";var f5U=f9Cm$[555616];I8Z.j$H();f5U+=p8U;f5U+=l4m;f5U+=O7d;var C5t=p26;C5t+=f9Cm$.t_T;C5t+=T52;var H2G=I1d;H2G+=K4G;H2G+=X2z;H2G+=A7I;var F9m=f9Cm$[228782];F9m+=E$W;var J1o=K18;J1o+=f9Cm$[481343];J1o+=K18;J1o+=f9Cm$.J4L;var z4_=A6T;z4_+=l6w;z4_+=K18;z4_+=I$m;var U_6=R6h;U_6+=f9Cm$.t_T;if(settings[U_6] !== table[C37]){return;}var api=new DataTable$3[s5A](settings);var containerNode=$(api[z6u](undefined)[z4_]());DataTable$3[j_o][J1o](api);side[P2$](containerNode[F9m](H2G))[P2$](containerNode[I$n](h5t))[C5t](containerNode[I$n](f5U));})[q7w]($[B05]({buttons:[],columns:[{data:conf[C4v][e9d],title:u09}],deferRender:X17,dom:n8e,language:{paginate:{next:f$B,previous:f5h},search:j_l,searchPlaceholder:G9y},lengthChange:h4R,select:{style:conf[Q$N]?Q9r:I1Z}},conf[R_t]));this[o3c](w2D,function(){var I8y="search";var b68="just";var e4n=O1b;e4n+=b68;var I1U=A6T;I1U+=k8K;I1U+=f9Cm$[481343];I1U+=f9Cm$.J3V;I8Z.j$H();if(dt[I8y]()){dt[I8y](j_l)[u8I]();}dt[I1U][e4n]();});dt[h8a](o6L,function(){I8Z.m_d();var B$L=J7A;B$L+=k_d;_triggerChange($(conf[F3v][B$L]()[O1T]()));});if(conf[u6x]){var T2c=f9Cm$.J3V;T2c+=n2z;T2c+=s6q;var j6S=f9Cm$[23424];j6S+=f9Cm$[481343];var B_y=w3$;B_y+=K18;B_y+=L2U;var E9O=f9Cm$.J4L;E9O+=j$2;E9O+=k_d;var J48=w3$;J48+=g8Y;J48+=A5h;conf[J48][E9O](dt);conf[B_y][j6S](T2c,function(e,json,data,action){var y61='refresh';var g6F="irst";var J$Y="_jumpTo";var I8T=J$Y;I8T+=y7j;I8T+=g6F;var A6Y=O0W;A6Y+=f9Cm$[23424];A6Y+=a6R;var u3k=f9Cm$.t_T;u3k+=f9Cm$[555616];u3k+=g8Y;if(action === c2D){var _loop_1=function(dp){dt[X9o](function(idx,d){return d === dp;})[j_o]();};for(var _i=C37,_a=json[I$4];_i < _a[B97];_i++){var dp=_a[_i];_loop_1(dp);}}else if(action === u3k || action === A6Y){_this[e3V](y61);}datatable[I8T](conf,_this);});}conf[F3v]=dt;datatable[T2j](conf,conf[E7n] || []);return {input:container,side:side};},disable:function(conf){var q1K=J0n;I8Z.m_d();q1K+=g7i;q1K+=L_M;q1K+=f9Cm$.J3V;var J6T=f9Cm$.e08;J6T+=B1d;J6T+=K18;conf[F3v][j_o][r_x](J6T);conf[F3v][q1K]()[O1T]()[X5f](M_p,n50);},dt:function(conf){I8Z.m_d();return conf[F3v];},enable:function(conf){var G0B="ock";var f3O=R6r;f3O+=G0B;var z0_=h44;z0_+=t_U;z0_+=f9Cm$.e08;z0_+=g5f;var F4s=f9Cm$[555616];F4s+=f9Cm$.J4L;var X91=f9Cm$.J3V;X91+=K18;X91+=L47;X91+=k_d;var u1p=Y1m;u1p+=Q6Q;u1p+=M$g;var V8u=f9Cm$[555616];V8u+=f9Cm$.J4L;conf[V8u][u1p][r_x](conf[Q$N]?Q9r:X91);I8Z.j$H();conf[F4s][d7J]()[O1T]()[X5f](z0_,f3O);},get:function(conf){var i4p="air";var K6r="pl";var J$q="separ";var c3c="nsP";var f1r="toAr";var a8w="uck";var b0o=J$q;b0o+=k6$;b0o+=f9Cm$[23424];b0o+=f9Cm$.l60;var T7O=f1r;T7O+=q7Q;T7O+=g5f;var T$k=B9J;T$k+=c3c;T$k+=i4p;var a$v=K6r;a$v+=a8w;var e2T=f9Cm$[555616];e2T+=k6$;e2T+=f9Cm$.e08;var H5D=f9Cm$.l60;H5D+=h98;I8Z.m_d();H5D+=f9Cm$.J3V;var rows=conf[F3v][H5D]({selected:X17})[e2T]()[a$v](conf[T$k][l8T])[T7O]();return conf[b0o] || !conf[Q$N]?rows[R5X](conf[v9Y] || U74):rows;},set:function(conf,val,localUpdate){var h0u="pli";var M9u="separato";var T5O="selec";var Y6a="nsPair";var Y2O="_jumpToFirst";var n2$="alue";var P2O="elect";var N9b=T5O;N9b+=f9Cm$.J4L;var o5g=f9Cm$[555616];o5g+=f9Cm$.J4L;var b3i=Q1L;b3i+=P2O;var Z$8=f9Cm$[555616];Z$8+=f9Cm$.J4L;I8Z.j$H();var t3o=I7r;t3o+=n2$;var c6r=B9J;c6r+=Y6a;var z2l=z5H;z2l+=J_S;z2l+=g5f;var r58=M9u;r58+=f9Cm$.l60;if(conf[Q$N] && conf[r58] && !Array[d2_](val)){var J8p=f9Cm$.J3V;J8p+=h0u;J8p+=f9Cm$.J4L;val=typeof val === t3X?val[J8p](conf[v9Y]):[];}else if(!Array[z2l](val)){val=[val];}var valueFn=dataGet(conf[c6r][t3o]);conf[Z$8][X9o]({selected:X17})[b3i]();conf[o5g][X9o](function(idx,data,node){I8Z.m_d();return val[x7i](valueFn(data)) !== -Y5Y;})[N9b]();datatable[Y2O](conf,this);if(!localUpdate){_triggerChange($(conf[F3v][z6u]()[O1T]()));}},tableClass:j_l,update:function(conf,options,append){var x$D="_lastSet";var c2T="_addO";var P10=f9Cm$[555616];P10+=f9Cm$.J4L;var b$7=c2T;I8Z.m_d();b$7+=B1d;b$7+=q5N;b$7+=f9c;datatable[b$7](conf,options,append);var lastSet=conf[x$D];if(lastSet !== undefined){datatable[D0u](conf,lastSet,X17);}_triggerChange($(conf[P10][z6u]()[O1T]()));}});var defaults={className:j_l,compare:B3c,data:j_l,def:j_l,entityDecode:X17,fieldInfo:j_l,getFormatter:B3c,id:j_l,label:j_l,labelInfo:j_l,message:j_l,multiEditable:X17,name:B3c,nullDefault:h4R,setFormatter:B3c,submit:X17,type:f4B};var DataTable$2=$[f9Cm$.E4X][f9Cm$.m96];var Field=(function(){var o0R="ompare";var C6A="rototy";var X7s="na";var w7o="proto";var A$B="Ids";var B21="pdate";var D1$="inputControl";var f7u="isMultiValu";var K8A="labelInfo";var Z7_="prototyp";var U8P="ototyp";var t9w="multiInfoShown";var v_S="_format";var t39="nullDefa";var e$A="sg";var O53="host";var g0i="ototy";var j2w="ul";var I7E="multiEditable";var o21='input, select, textarea';var S2V="multiReturn";var j5s="otype";var d9B="oty";var B9U="prot";var e9V="mittabl";var U5f="slideUp";var N6W="isMultiValue";var K3K="Get";var e5z="fieldInfo";var i4y="_typ";var w1m="formatters";var n7v="_msg";var t3j="rot";var q0I="otyp";var c77="multiValues";var f8m="defaults";var r71="rNod";var g8A="eFn";var h1x="ces";var a5D="multiRes";var J4P="_err";var F4O=J4P;F4O+=f9Cm$[23424];F4O+=r71;F4O+=f9Cm$.t_T;var C0j=i4y;C0j+=g8A;var q$E=w7o;q$E+=c11;var s2D=j7g;s2D+=D6p;s2D+=e$A;var m_w=w7o;m_w+=c11;var l1o=k18;l1o+=e9V;l1o+=f9Cm$.t_T;var u8V=G2J;u8V+=f9Cm$.J4L;u8V+=d9B;u8V+=p7O;var y3G=C95;y3G+=A$B;var R$y=G2J;R$y+=V9Y;var r0c=B1d;r0c+=C6A;r0c+=p7O;var O$G=u6g;O$G+=O6L;var u0T=q9Z;u0T+=o0R;var u5X=Z7_;u5X+=f9Cm$.t_T;var Z_R=f9Cm$.Y87;Z_R+=B21;var B$t=u6g;B$t+=U8P;B$t+=f9Cm$.t_T;var h96=f9Cm$.J3V;h96+=u1J;h96+=H3b;var k25=f9Cm$.J3V;k25+=f9Cm$.t_T;k25+=f9Cm$.J4L;var s8e=u6g;s8e+=f9Cm$[23424];s8e+=h1x;s8e+=G7W;var u1q=t39;u1q+=j2w;u1q+=f9Cm$.J4L;var o2K=B1d;o2K+=t3j;o2K+=q0I;o2K+=f9Cm$.t_T;var Q7l=L6$;Q7l+=f9Cm$.t_T;var O_L=C2U;O_L+=C2h;O_L+=f9Cm$.t_T;O_L+=f9Cm$.J4L;var C6$=a5D;C6$+=f9Cm$.J4L;C6$+=A5h;C6$+=f9Cm$.t_T;var w$c=C95;w$c+=K3K;var j2b=u6g;j2b+=f9Cm$[23424];j2b+=f9Cm$.J4L;j2b+=j5s;var L6u=u6g;L6u+=g0i;L6u+=p7O;var r2F=u6g;r2F+=e0Y;r2F+=q0I;r2F+=f9Cm$.t_T;var K3l=B9U;K3l+=e0Y;K3l+=g5f;K3l+=p7O;var u3U=u6g;u3U+=g0i;u3U+=p7O;var t9Z=f7u;t9Z+=f9Cm$.t_T;var R_h=w7o;R_h+=q_4;R_h+=B1d;R_h+=f9Cm$.t_T;var e4H=n3_;e4H+=j$2;e4H+=Q6Q;e4H+=w3$;var X18=G2J;X18+=V9Y;var B1C=B9U;B1C+=f9Cm$[23424];B1C+=c11;var m4Z=h44;m4Z+=f9Cm$.J3V;m4Z+=l27;var p4A=R3S;p4A+=f9Cm$[228782];function Field(options,classes,host){var c2A='" for="';var o0e="<spa";var v1W="ypes";var J73="ulti-";var U6V="itle";var f9b="iv data";var A1Y="Val";var U5B="g-info";var J7P="Prefi";var S18="ocessing\" class=\"";var Q8F="sg-error\" class=\"";var I4J='<label data-dte-e="label" class="';var N9N="internal";var P6B="sg-erro";var R9T="restore";I8Z.j$H();var P8P="<div data-dte-e=\"inp";var R1g="n ";var N59="efix";var l8A="ldT";var X3d="ault";var o4d="msg-mu";var w4d="msg-multi\" class=\"";var W1Z="<div data-";var f5N='msg-label';var x64='msg-error';var P_x="ms";var L7q="From";var g$U='Error adding field - unknown field type ';var k9n="multiInfo";var B6M="t-co";var M7l="msg-";var p3X="data-dte-e=\"multi-info\" class=\"";var x1p='input-control';var o0X='msg-message';var W9W='field-processing';var A6S="valT";var c2z="oData";var o5N="msg-info\" class=\"";var A5r="-dte-e=\"";var z_3='<div data-dte-e="msg-label" class="';var Y39="I18n";var d9E="<div data-dte-e=\"field-pr";var B$U="estore";var Y78="n>";var l6U="spa";var a$C="te-e=\"m";var e6i="trol";var x6j='<div data-dte-e="msg-message" class="';var E$L="ltiR";var I8n="dte-e=\"input-control\" class=\"";var Q9x="ut\" class=\"";var Q6b="<div data-dte-e=\"";var w_U="side";var I5E="></";var h9Q='<div data-dte-e="multi-value" class="';var D7Z="data-d";var q8a="E_F";var g1n=f9Cm$.t_T;g1n+=o$n;g1n+=s$2;var a0q=l_m;a0q+=e3G;var S63=f9Cm$[23424];S63+=f9Cm$[481343];var P16=f9Cm$[555616];P16+=f9Cm$[23424];P16+=D6p;var f$u=f9Cm$[23424];f$u+=f9Cm$[481343];var a2a=o4d;a2a+=Q6Q;a2a+=B4Y;var t8D=D6p;t8D+=J73;t8D+=K18;t8D+=q9g;var U5z=M7l;U5z+=K18;U5z+=q9g;var n0Y=o9d;n0Y+=D6p;var p0B=q4o;p0B+=n_a;var F2$=k8A;F2$+=j0k;F2$+=T5I;var T3t=h7u;T3t+=B12;T3t+=G7W;var O1y=d9E;O1y+=S18;var p7a=Z4$;p7a+=h8B;p7a+=i$Z;var g9f=k8A;g9f+=T5B;var k3r=y9e;k3r+=i$Z;var R1v=P_x;R1v+=U5B;var m1C=Q6b;m1C+=o5N;var J77=Z4$;J77+=E3u;var G7q=y9e;G7q+=I5E;G7q+=E3u;var Z9B=D6p;Z9B+=P6B;Z9B+=f9Cm$.l60;var Y6j=d9M;Y6j+=D7Z;Y6j+=a$C;Y6j+=Q8F;var C8v=k8A;C8v+=R$r;C8v+=h44;C8v+=T5I;var I97=y9e;I97+=i$Z;var l9K=M6F;l9K+=E$L;l9K+=B$U;var u1L=q7L;u1L+=f9b;u1L+=A5r;u1L+=w4d;var d35=Z4$;d35+=h44;d35+=T5I;var g7t=Z4$;g7t+=l6U;g7t+=Y78;var Q7p=B96;Q7p+=f9Cm$[228782];Q7p+=f9Cm$[23424];var G8G=y9e;G8G+=i$Z;var k7N=o0e;k7N+=R1g;k7N+=p3X;var y15=f9Cm$.J4L;y15+=U6V;var O5I=y9e;O5I+=i$Z;var c57=C95;c57+=A1Y;c57+=X3P;var L_7=W1Z;L_7+=I8n;var S7Y=P8P;S7Y+=Q9x;var d5V=Z4$;d5V+=f9Cm$[555616];d5V+=x2W;var Y$$=Q6Q;Y$$+=f9Cm$.e08;Y$$+=h2q;Y$$+=Q6Q;var P7l=y9e;P7l+=i$Z;var I8Q=K18;I8Q+=f9Cm$[555616];var S7o=X7s;S7o+=W1R;S7o+=J7P;S7o+=G1H;var p03=C18;p03+=f9Cm$.t_T;var T6_=c11;T6_+=w_7;T6_+=f9Cm$.l60;T6_+=N59;var Q6p=p5y;Q6p+=C$g;Q6p+=B1d;Q6p+=F7I;var R3d=f9Cm$[555616];R3d+=q94;var q7W=A6S;q7W+=c2z;var L_5=P1N;L_5+=L7q;L_5+=l2H;L_5+=Q3e;var u$f=F3o;u$f+=f9Cm$.t_T;u$f+=l8A;u$f+=v1W;var k4d=f9Cm$.J4L;k4d+=g5f;k4d+=B1d;k4d+=f9Cm$.t_T;var j_j=f9Cm$[555616];j_j+=o2Y;j_j+=X3d;j_j+=f9Cm$.J3V;var j7i=N9N;j7i+=Y39;var that=this;var multiI18n=host[j7i]()[C95];var opts=$[d7H](X17,{},Field[j_j],options);if(!Editor[t3_][opts[k4d]]){throw new Error(g$U + opts[c11]);}this[f9Cm$.J3V]={classes:classes,host:host,multiIds:[],multiValue:h4R,multiValues:{},name:opts[h2d],opts:opts,processing:h4R,type:Editor[u$f][opts[c11]]};if(!opts[U7A]){var F4H=X7s;F4H+=W1R;var d1E=B0_;d1E+=q8a;d1E+=g2a;d1E+=j7g;opts[U7A]=d1E + opts[F4H];}if(opts[I$4] === j_l){var D3n=f9Cm$[481343];D3n+=f9Cm$.e08;D3n+=W1R;opts[I$4]=opts[D3n];}this[L_5]=function(d){var D1_=C6T;D1_+=f9Cm$[23424];D1_+=f9Cm$.l60;var P96=f9Cm$[555616];P96+=f9Cm$.e08;P96+=f9Cm$.J4L;P96+=f9Cm$.e08;return dataGet(opts[P96])(d,D1_);};this[q7W]=dataSet(opts[R3d]);var template=$(v4Q + classes[Q6p] + E$d + classes[T6_] + opts[p03] + E$d + classes[S7o] + opts[h2d] + E$d + opts[P03] + x$l + I4J + classes[e9d] + c2A + Editor[L2y](opts[I8Q]) + P7l + opts[Y$$] + z_3 + classes[f5N] + x$l + opts[K8A] + d5V + t7l + S7Y + classes[S8e] + x$l + L_7 + classes[D1$] + g42 + h9Q + classes[c57] + O5I + multiI18n[y15] + k7N + classes[k9n] + G8G + multiI18n[Q7p] + g7t + d35 + u1L + classes[l9K] + I97 + multiI18n[R9T] + C8v + Y6j + classes[Z9B] + G7q + x6j + classes[o0X] + x$l + opts[c1_] + J77 + m1C + classes[R1v] + k3r + opts[e5z] + g9f + p7a + O1y + classes[T3t] + R4f + F2$);var input=this[Y_T](p0B,opts);var side=B3c;if(input && input[w_U]){var W10=f9Cm$.J3V;W10+=K18;W10+=f9Cm$[555616];W10+=f9Cm$.t_T;side=input[W10];input=input[S8e];}if(input !== B3c){var A5p=M_q;A5p+=B6M;A5p+=f9Cm$[481343];A5p+=e6i;el(A5p,template)[g9P](input);}else {template[X5f](M_p,n50);}this[n0Y]={container:template,fieldError:el(x64,template),fieldInfo:el(U5z,template),fieldMessage:el(o0X,template),inputControl:el(x1p,template),label:el(B6L,template)[P2$](side),labelInfo:el(f5N,template),multi:el(Y9J,template),multiInfo:el(t8D,template),multiReturn:el(a2a,template),processing:el(W9W,template)};this[X6t][C95][f$u](F7k,function(){var U5H='readonly';var G7Z=f9Cm$[23424];G7Z+=g$5;if(that[f9Cm$.J3V][G7Z][I7E] && !template[P6T](classes[n0f]) && opts[c11] !== U5H){var j$h=f9Cm$[228782];j$h+=f9Cm$[23424];j$h+=e4v;j$h+=f9Cm$.J3V;that[P1N](j_l);that[j$h]();}});this[P16][S2V][S63](a0q,function(){var z1A="ultiR";var X_X=D6p;X_X+=z1A;I8Z.j$H();X_X+=B12;X_X+=S3f;that[X_X]();});$[g1n](this[f9Cm$.J3V][c11],function(name,fn){if(typeof fn === t0t && that[name] === undefined){that[name]=function(){I8Z.j$H();var u37=q9Z;u37+=f9Cm$.e08;u37+=Q6Q;u37+=Q6Q;var args=Array[P99][k3q][u37](arguments);args[d9t](name);var ret=that[Y_T][D08](that,args);return ret === undefined?that:ret;};}});}Field[P99][p4A]=function(set){var L5C='default';var k2i=J_q;k2i+=f9Cm$.E9M;var opts=this[f9Cm$.J3V][k2i];if(set === undefined){var z58=e96;z58+=j8w;z58+=f9Cm$[23424];z58+=f9Cm$[481343];var def=opts[L5C] !== undefined?opts[L5C]:opts[Z87];return typeof def === z58?def():def;}opts[Z87]=set;return this;};Field[P99][m4Z]=function(){var z_4="tain";var a4e=E57;a4e+=z_4;a4e+=f9Cm$.t_T;a4e+=f9Cm$.l60;var y6m=f9Cm$[555616];y6m+=f9Cm$[23424];y6m+=D6p;this[y6m][a4e][k1$](this[f9Cm$.J3V][A2x][n0f]);this[Y_T](O6K);return this;};Field[B1C][n5R]=function(){var m8M=b7s;m8M+=I8$;var t7G=J0n;t7G+=f9Cm$[23424];t7G+=f9Cm$[555616];t7G+=g5f;var q6i=D6_;q6i+=w2Q;q6i+=f9Cm$.J4L;q6i+=f9Cm$.J3V;var f7B=f9Cm$[555616];f7B+=f9Cm$[23424];I8Z.m_d();f7B+=D6p;var container=this[f7B][O1T];return container[q6i](t7G)[B97] && container[X5f](m8M) !== n50?X17:h4R;};Field[P99][y9q]=function(toggle){var g7Z="_ty";var H4T="aine";var Q71='enable';var I8V=g7Z;I8V+=p7O;I8V+=y7j;I8V+=f9Cm$[481343];var h0B=Q06;h0B+=a6R;h0B+=E86;var Z7C=E57;Z7C+=f9Cm$.J4L;Z7C+=H4T;Z7C+=f9Cm$.l60;var a9B=f9Cm$[555616];a9B+=u8G;if(toggle === void C37){toggle=X17;}if(toggle === h4R){var a3e=f9Cm$[555616];a3e+=K_b;a3e+=f9Cm$.e08;a3e+=V5s;return this[a3e]();}this[a9B][Z7C][h0B](this[f9Cm$.J3V][A2x][n0f]);this[I8V](Q71);return this;};Field[X18][e4H]=function(){var x3K=h44;I8Z.m_d();x3K+=f9Cm$.J3V;x3K+=Z_K;var z8_=r8u;z8_+=f9Cm$.J3V;z8_+=f9Cm$.t_T;z8_+=f9Cm$.J3V;var T_p=o9d;T_p+=D6p;return this[T_p][O1T][P6T](this[f9Cm$.J3V][z8_][x3K]) === h4R;};Field[R_h][h8G]=function(msg,fn){var Y9c="oveC";var Y6Z='errorMessage';var G1n=o9d;G1n+=D6p;var b7W=t9a;b7W+=f9Cm$.J3V;b7W+=h7H;var I7X=q9Z;I7X+=Q6Q;I7X+=S$6;I7X+=B12;var classes=this[f9Cm$.J3V][I7X];if(msg){var X03=O1b;X03+=f9Cm$[555616];X03+=B7X;X03+=S$6;var f5z=q9Z;f5z+=a3Y;f5z+=i0g;f5z+=I$m;this[X6t][f5z][X03](classes[h8G]);}else {var x1d=F7I;x1d+=h5o;var f$_=O0W;f$_+=Y9c;f$_+=S1i;f$_+=i3q;this[X6t][O1T][f$_](classes[x1d]);}I8Z.m_d();this[Y_T](Y6Z,msg);return this[b7W](this[G1n][F81],msg,fn);};Field[P99][e5z]=function(msg){return this[n7v](this[X6t][e5z],msg);};Field[P99][t9Z]=function(){var T3d="lue";var z3e="tiIds";var l2A="iV";var S3E=C2U;S3E+=z3e;var T$F=U3L;T$F+=l2A;T$F+=f9Cm$.e08;T$F+=T3d;return this[f9Cm$.J3V][T$F] && this[f9Cm$.J3V][S3E][B97] !== Y5Y;};Field[u3U][G9w]=function(){var W_h="ainer";var y17=t3Z;y17+=f9Cm$.l60;I8Z.m_d();var S9$=Y$v;S9$+=W_h;return this[X6t][S9$][P6T](this[f9Cm$.J3V][A2x][y17]);};Field[P99][S8e]=function(){var W5G=B96;I8Z.m_d();W5G+=p_t;var y6e=K18;y6e+=f9Cm$[481343];y6e+=B1d;y6e+=g7i;return this[f9Cm$.J3V][c11][y6e]?this[Y_T](W5G):$(o21,this[X6t][O1T]);};Field[K3l][G0n]=function(){var G9U=f9Cm$[228782];G9U+=f9Cm$[23424];G9U+=e4v;G9U+=f9Cm$.J3V;var b0x=q_4;b0x+=p7O;if(this[f9Cm$.J3V][b0x][G9U]){this[Y_T](F2a);}else {$(o21,this[X6t][O1T])[G0n]();}return this;};Field[P99][W7y]=function(){var W4i="_for";var e6$="getFormat";var g7_=e6$;g7_+=f9Cm$.J4L;g7_+=F7I;var q1q=f9Cm$[23424];q1q+=B1d;q1q+=f9Cm$.J4L;q1q+=f9Cm$.J3V;var H$m=h7H;H$m+=f9Cm$.t_T;H$m+=f9Cm$.J4L;var I6r=i4y;I6r+=g8A;var a1L=W4i;a1L+=Y0o;if(this[N6W]()){return undefined;}return this[a1L](this[I6r](H$m),this[f9Cm$.J3V][q1q][g7_]);};Field[r2F][J7j]=function(animate){var j_V="ispl";I8Z.j$H();var P7j=f9Cm$[228782];P7j+=f9Cm$[481343];var r1n=f9Cm$[555616];r1n+=j_V;r1n+=n7x;var q0y=f9Cm$[555616];q0y+=f9Cm$[23424];q0y+=D6p;var el=this[q0y][O1T];if(animate === undefined){animate=X17;}if(this[f9Cm$.J3V][O53][r1n]() && animate && $[P7j][U5f]){el[U5f]();}else {var R24=k14;R24+=f9Cm$[481343];R24+=f9Cm$.t_T;var l3W=q9Z;l3W+=f9Cm$.J3V;l3W+=f9Cm$.J3V;el[l3W](M_p,R24);}return this;};Field[P99][e9d]=function(str){var a9p="tm";var R35=s_E;R35+=j1L;I8Z.m_d();var D74=f9Cm$[555616];D74+=f9Cm$[23424];D74+=D6p;var label=this[X6t][e9d];var labelInfo=this[D74][K8A][R35]();if(str === undefined){var C7B=s$2;C7B+=a9p;C7B+=Q6Q;return label[C7B]();}label[n8n](str);label[P2$](labelInfo);return this;};Field[L6u][K8A]=function(msg){var H_$="labe";var N5D=H_$;N5D+=Q6Q;I8Z.j$H();N5D+=Y5K;N5D+=f9Cm$[23424];return this[n7v](this[X6t][N5D],msg);};Field[j2b][c1_]=function(msg,fn){var U9u="Message";var B91=f9Cm$[228782];B91+=s4W;B91+=e30;I8Z.m_d();B91+=U9u;var r$e=o9d;r$e+=D6p;return this[n7v](this[r$e][B91],msg,fn);};Field[P99][w$c]=function(id){var value;I8Z.j$H();var multiValues=this[f9Cm$.J3V][c77];var multiIds=this[f9Cm$.J3V][a4h];var isMultiValue=this[N6W]();if(id === undefined){var G5C=D3W;G5C+=i1I;G5C+=s$2;var O7a=I7r;O7a+=M_7;var fieldVal=this[O7a]();value={};for(var _i=C37,multiIds_1=multiIds;_i < multiIds_1[G5C];_i++){var multiId=multiIds_1[_i];value[multiId]=isMultiValue?multiValues[multiId]:fieldVal;}}else if(isMultiValue){value=multiValues[id];}else {value=this[P1N]();}return value;};Field[P99][C6$]=function(){var a4X="multiValue";this[f9Cm$.J3V][a4X]=X17;this[H3$]();};Field[P99][O_L]=function(id,val,recalc){var u6P="eck";var B$S="Value";var H1S="sPlainObject";var a2v="multiI";var B74="ueCh";var Y$m="ltiVal";var w83="_mu";var P0J=C95;P0J+=B$S;var T_t=K18;T_t+=H1S;var T2V=a2v;T2V+=X5O;I8Z.j$H();if(recalc === void C37){recalc=X17;}var that=this;var multiValues=this[f9Cm$.J3V][c77];var multiIds=this[f9Cm$.J3V][T2V];if(val === undefined){val=id;id=undefined;}var set=function(idSrc,valIn){var g$2="setFormatter";var D1c=v7I;D1c+=M55;D1c+=f9Cm$.J4L;if($[G_5](idSrc,multiIds) === -Y5Y){var u$9=B1d;u$9+=f9Cm$.Y87;u$9+=f9Cm$.J3V;u$9+=s$2;multiIds[u$9](idSrc);}multiValues[idSrc]=that[D1c](valIn,that[f9Cm$.J3V][w3u][g$2]);};if($[T_t](val) && id === undefined){var P88=q4u;P88+=j1L;$[P88](val,function(idSrc,innerVal){I8Z.j$H();set(idSrc,innerVal);});}else if(id === undefined){$[a8N](multiIds,function(i,idSrc){I8Z.j$H();set(idSrc,val);});}else {set(id,val);}this[f9Cm$.J3V][P0J]=X17;if(recalc){var m0w=w83;m0w+=Y$m;m0w+=B74;m0w+=u6P;this[m0w]();}return this;};Field[P99][h2d]=function(){var H_g=X7s;I8Z.m_d();H_g+=D6p;H_g+=f9Cm$.t_T;return this[f9Cm$.J3V][w3u][H_g];};Field[P99][Q7l]=function(){var A6I=f9Cm$[555616];A6I+=f9Cm$[23424];A6I+=D6p;return this[A6I][O1T][C37];};Field[o2K][u1q]=function(){var d1e="nullDefault";return this[f9Cm$.J3V][w3u][d1e];};Field[P99][s8e]=function(set){var s8w='processing-field';var B$M="ernalE";var T32=K18;T32+=l1E;T32+=B$M;T32+=e2C;var m3b=k14;m3b+=L$V;var b7L=s8o;b7L+=n7x;var D6X=q9Z;D6X+=f9Cm$.J3V;D6X+=f9Cm$.J3V;var U4N=G2J;U4N+=q9Z;U4N+=B12;U4N+=G7W;if(set === undefined){return this[f9Cm$.J3V][T0W];}this[X6t][U4N][D6X](b7L,set?M20:m3b);I8Z.m_d();this[f9Cm$.J3V][T0W]=set;this[f9Cm$.J3V][O53][T32](s8w,[set]);return this;};Field[P99][k25]=function(val,multiCheck){var C$V="ormatte";var G6K='set';var E3R="_multiVal";var d_V="ueChe";var G2f="tiValu";var h42="etF";var X6o="entityDecode";var p82=C2U;p82+=G2f;p82+=f9Cm$.t_T;if(multiCheck === void C37){multiCheck=X17;}var decodeFn=function(d){var E5u='"';var P4t="lac";var m5n='\n';var Z3a='£';var M97='\'';var U5E=A8G;U5E+=i8m;I8Z.m_d();var q3j=f9Cm$.l60;q3j+=P9K;q3j+=P4t;q3j+=f9Cm$.t_T;var P$t=K2g;P$t+=H5i;var i5a=f9Cm$.l60;i5a+=d9i;i5a+=f9Cm$.e08;i5a+=z8x;return typeof d !== t3X?d:d[d9I](/&gt;/g,f$B)[d9I](/&lt;/g,f5h)[i5a](/&amp;/g,s2o)[P$t](/&quot;/g,E5u)[d9I](/&#163;/g,Z3a)[q3j](/&#0?39;/g,M97)[U5E](/&#0?10;/g,m5n);};this[f9Cm$.J3V][p82]=h4R;var decode=this[f9Cm$.J3V][w3u][X6o];if(decode === undefined || decode === X17){if(Array[d2_](val)){var d8X=k_d;d8X+=y8m;for(var i=C37,ien=val[d8X];i < ien;i++){val[i]=decodeFn(val[i]);}}else {val=decodeFn(val);}}if(multiCheck === X17){var O0I=E3R;O0I+=d_V;O0I+=q9Z;O0I+=e3G;var W4I=f9Cm$.J3V;W4I+=h42;W4I+=C$V;W4I+=f9Cm$.l60;var q4v=f9Cm$[23424];q4v+=B1d;q4v+=f9Cm$.E9M;val=this[v_S](val,this[f9Cm$.J3V][q4v][W4I]);this[Y_T](G6K,val);this[O0I]();}else {this[Y_T](G6K,val);}return this;};Field[P99][h96]=function(animate,toggle){var f93="eD";var H8_="slid";var X4u=H8_;X4u+=f93;X4u+=n2J;I8Z.j$H();var g5b=u1J;g5b+=U5a;var b15=f9Cm$[555616];b15+=f9Cm$[23424];b15+=D6p;if(animate === void C37){animate=X17;}if(toggle === void C37){toggle=X17;}if(toggle === h4R){return this[J7j](animate);}var el=this[b15][O1T];if(this[f9Cm$.J3V][g5b][G7z]() && animate && $[f9Cm$.E4X][X4u]){var X9b=H8_;X9b+=f93;X9b+=n2J;el[X9b]();}else {var d2z=q9Z;d2z+=f9Cm$.J3V;d2z+=f9Cm$.J3V;el[d2z](M_p,j_l);;}return this;};Field[B$t][Z_R]=function(options,append){var b8V="yp";if(append === void C37){append=h4R;}if(this[f9Cm$.J3V][c11][T0s]){var k02=j7q;k02+=f9Cm$[555616];k02+=o2e;var f0E=b7w;f0E+=b8V;f0E+=g8A;this[f0E](k02,options,append);}return this;};Field[P99][P1N]=function(val){var n8W=h7H;n8W+=C6V;return val === undefined?this[n8W]():this[D0u](val);};Field[u5X][u0T]=function(value,original){var h66="compare";var Y8H=J_q;Y8H+=f9Cm$.E9M;I8Z.m_d();var compare=this[f9Cm$.J3V][Y8H][h66] || deepCompare;return compare(value,original);};Field[P99][s3Y]=function(){return this[f9Cm$.J3V][w3u][I$4];};Field[O$G][J8j]=function(){var O21="_typeF";var b7S="est";var U9i=f9Cm$[555616];U9i+=b7S;U9i+=a_T;U9i+=g5f;var A7l=O21;A7l+=f9Cm$[481343];this[X6t][O1T][h6J]();this[A7l](U9i);return this;};Field[r0c][I7E]=function(){var L3r="iEdi";I8Z.m_d();var c9v=U3L;c9v+=L3r;c9v+=z6u;return this[f9Cm$.J3V][w3u][c9v];};Field[R$y][y3G]=function(){I8Z.j$H();return this[f9Cm$.J3V][a4h];};Field[u8V][t9w]=function(show){var s$y="iIn";I8Z.m_d();var X2B=R6r;X2B+=f9Cm$[23424];X2B+=q9Z;X2B+=e3G;var a3R=q9Z;a3R+=f9Cm$.J3V;a3R+=f9Cm$.J3V;var N7e=U3L;N7e+=s$y;N7e+=f9Cm$[228782];N7e+=f9Cm$[23424];var B7C=f9Cm$[555616];B7C+=f9Cm$[23424];B7C+=D6p;this[B7C][N7e][a3R]({display:show?X2B:n50});};Field[P99][e3n]=function(){I8Z.j$H();this[f9Cm$.J3V][a4h]=[];this[f9Cm$.J3V][c77]={};};Field[P99][l1o]=function(){var d3w=f9Cm$[23424];d3w+=g$5;return this[f9Cm$.J3V][d3w][n3l];};Field[m_w][s2D]=function(el,msg,fn){var L8O="sible";var l_h="nternalSet";I8Z.j$H();var x_m="slideDown";var l3l="pare";var u4G="ima";var v4a=f9Cm$.e08;v4a+=f9Cm$[481343];v4a+=u4G;v4a+=U58;var C29=X2s;C29+=K18;C29+=L8O;var Q7j=K18;Q7j+=f9Cm$.J3V;var b93=l3l;b93+=l1E;if(msg === undefined){var r5F=h0z;r5F+=Q6Q;return el[r5F]();}if(typeof msg === t0t){var G01=Y4w;G01+=f9Cm$.t_T;var j3B=K18;j3B+=l_h;j3B+=z7B;var editor=this[f9Cm$.J3V][O53];msg=msg(editor,new DataTable$2[s5A](editor[j3B]()[G01]));}if(el[b93]()[Q7j](C29) && $[f9Cm$.E4X][v4a]){el[n8n](msg);if(msg){el[x_m](fn);;}else {el[U5f](fn);}}else {var p0U=k14;p0U+=L$V;var V2e=f9Cm$[555616];V2e+=K18;V2e+=V1c;V2e+=I8$;var S9D=q9Z;S9D+=i3q;el[n8n](msg || j_l)[S9D](V2e,msg?M20:p0U);if(fn){fn();}}return this;};Field[q$E][H3$]=function(){var o2r="lMultiInfo";var Y1$="ulti";var e$p="multiValu";var x_u="rna";var f7t="ltiVa";var B7o="inte";var p3k="toggleClas";var C_K="lues";var g6T="lI18";var y9J="noMult";var u4D="interna";var a1A="ltiNoEdit";var K_o=u4D;K_o+=o2r;var h4E=s$2;h4E+=f9Cm$[23424];h4E+=f9Cm$.J3V;h4E+=f9Cm$.J4L;var U2n=D6p;U2n+=f9Cm$.Y87;U2n+=a1A;var D7v=p3k;D7v+=f9Cm$.J3V;var N0h=C2U;N0h+=B4Y;var G43=y9J;G43+=K18;var v4K=I_i;v4K+=f9Cm$[23424];var S$r=D6p;S$r+=Y1$;var x6K=B7o;x6K+=x_u;x6K+=g6T;x6K+=f9Cm$[481343];var F_R=f9Cm$[481343];F_R+=f9Cm$[23424];F_R+=f9Cm$[481343];F_R+=f9Cm$.t_T;var S$4=q9Z;S$4+=i3q;var L6g=C95;L6g+=x1X;var r$0=W6q;r$0+=f9Cm$.J3V;var s8M=e$p;s8M+=f9Cm$.t_T;var v$X=M6F;v$X+=f7t;v$X+=C_K;var last;var ids=this[f9Cm$.J3V][a4h];var values=this[f9Cm$.J3V][v$X];var isMultiValue=this[f9Cm$.J3V][s8M];var isMultiEditable=this[f9Cm$.J3V][r$0][L6g];var val;var different=h4R;if(ids){var t$m=k_d;t$m+=y8m;for(var i=C37;i < ids[t$m];i++){val=values[ids[i]];if(i > C37 && !deepCompare(val,last)){different=X17;break;}last=val;}}if(different && isMultiValue || !isMultiEditable && this[N6W]()){var N1v=R6r;N1v+=f9Cm$[23424];N1v+=q9Z;N1v+=e3G;var h3S=q9Z;h3S+=f9Cm$.J3V;h3S+=f9Cm$.J3V;var Y7h=f9Cm$[481343];Y7h+=f9Cm$[23424];Y7h+=f9Cm$[481343];Y7h+=f9Cm$.t_T;var m8v=q9Z;m8v+=f9Cm$.J3V;m8v+=f9Cm$.J3V;var B3e=S8e;B3e+=s0$;var V0q=f9Cm$[555616];V0q+=f9Cm$[23424];V0q+=D6p;this[V0q][B3e][m8v]({display:Y7h});this[X6t][C95][h3S]({display:N1v});}else {var v9z=k14;v9z+=L$V;var E8m=M6F;E8m+=e$q;var r_X=q9Z;r_X+=f9Cm$.J3V;r_X+=f9Cm$.J3V;this[X6t][D1$][r_X]({display:M20});this[X6t][E8m][X5f]({display:v9z});if(isMultiValue && !different){var B1k=f9Cm$.J3V;B1k+=f9Cm$.t_T;B1k+=f9Cm$.J4L;this[B1k](last,h4R);}}this[X6t][S2V][S$4]({display:ids && ids[B97] > Y5Y && different && !isMultiValue?M20:F_R});var i18n=this[f9Cm$.J3V][O53][x6K]()[S$r];this[X6t][v4K][n8n](isMultiEditable?i18n[c0T]:i18n[G43]);this[X6t][N0h][D7v](this[f9Cm$.J3V][A2x][U2n],!isMultiEditable);this[f9Cm$.J3V][h4E][K_o]();return X17;};Field[P99][C0j]=function(name){var K$O=f9Cm$[23424];K$O+=B1d;K$O+=f9Cm$.J4L;K$O+=f9Cm$.J3V;var W7U=k_d;W7U+=y8m;var args=[];for(var _i=Y5Y;_i < arguments[W7U];_i++){args[_i - Y5Y]=arguments[_i];}args[d9t](this[f9Cm$.J3V][K$O]);var fn=this[f9Cm$.J3V][c11][name];I8Z.j$H();if(fn){return fn[D08](this[f9Cm$.J3V][O53],args);}};Field[P99][F4O]=function(){var R9m=j2Z;R9m+=A5h;I8Z.j$H();var Q6G=f9Cm$[555616];Q6G+=f9Cm$[23424];Q6G+=D6p;return this[Q6G][R9m];};Field[P99][v_S]=function(val,formatter){if(formatter){var N3A=q9Z;N3A+=f9Cm$.e08;N3A+=Q6Q;N3A+=Q6Q;if(Array[d2_](formatter)){var R8D=f9Cm$.J3V;R8D+=i8Z;R8D+=f9Cm$[228782];R8D+=f9Cm$.J4L;var M0M=f9Cm$.J3V;M0M+=k1J;M0M+=z8x;var args=formatter[M0M]();var name_1=args[R8D]();formatter=Field[w1m][name_1][D08](this,args);}return formatter[N3A](this[f9Cm$.J3V][O53],val,this);}return val;};Field[f8m]=defaults;Field[w1m]={};return Field;})();var button={action:B3c,className:B3c,tabIndex:C37,text:B3c};var displayController={close:function(){},init:function(){},node:function(){},open:function(){}};var DataTable$1=$[f9Cm$.E4X][z0W];var apiRegister=DataTable$1[s5A][N1$];function _getInst(api){var u_c="context";var u_r="_editor";var h9f=f9Cm$[23424];h9f+=h3G;h9f+=f9Cm$[481343];h9f+=g8Y;var ctx=api[u_c][C37];return ctx[h9f][Z9K] || ctx[u_r];}function _setBasic(inst,opts,type,plural){var u1h="fir";var t$M=/%d/;var F9r=W1R;F9r+=i3q;I8Z.j$H();F9r+=f9Cm$.e08;F9r+=P9U;if(!opts){opts={};}if(opts[d7J] === undefined){var A4G=q1i;A4G+=I4O;A4G+=W4z;opts[A4G]=f9B;}if(opts[M3D] === undefined){opts[M3D]=inst[B5C][type][M3D];}if(opts[F9r] === undefined){if(type === m8Z){var n88=D6p;n88+=z_k;n88+=D2Y;var V2f=E57;V2f+=u1h;V2f+=D6p;var confirm_1=inst[B5C][type][V2f];opts[n88]=plural !== Y5Y?confirm_1[j7g][d9I](t$M,plural):confirm_1[T1p];}else {opts[c1_]=j_l;}}return opts;}apiRegister(Z7$,function(){I8Z.m_d();return _getInst(this);});apiRegister(c7B,function(opts){var q80=q9Z;q80+=g7G;q80+=f9Cm$.J4L;q80+=f9Cm$.t_T;var z_q=q9Z;z_q+=f9Cm$.l60;z_q+=f9Cm$.t_T;I8Z.j$H();z_q+=o2e;var inst=_getInst(this);inst[z_q](_setBasic(inst,opts,q80));return this;});apiRegister(a8U,function(opts){var k0x=G3n;k0x+=f9Cm$.J4L;var inst=_getInst(this);inst[C6T](this[C37][C37],_setBasic(inst,opts,k0x));return this;});apiRegister(N6r,function(opts){var G3_=f9Cm$.t_T;G3_+=f9Cm$[555616];G3_+=K18;G3_+=f9Cm$.J4L;var G$n=f9Cm$.t_T;G$n+=h44;G$n+=f9Cm$.J4L;var inst=_getInst(this);inst[G$n](this[C37],_setBasic(inst,opts,G3_));return this;});apiRegister(N5Q,function(opts){var inst=_getInst(this);inst[h6J](this[C37][C37],_setBasic(inst,opts,m8Z,Y5Y));return this;});apiRegister(l4f,function(opts){var A5z=O4Y;A5z+=R0T;var inst=_getInst(this);inst[h6J](this[C37],_setBasic(inst,opts,A5z,this[C37][B97]));return this;});apiRegister(D5e,function(type,opts){var F9Z=M4P;F9Z+=B96;F9Z+=F6K;I8Z.j$H();F9Z+=l$O;if(!type){var t0n=K18;t0n+=I$Q;type=t0n;}else if($[F9Z](type)){opts=type;type=O8g;}_getInst(this)[type](this[C37][C37],opts);return this;});apiRegister(E_L,function(opts){_getInst(this)[m0D](this[C37],opts);return this;});apiRegister(x6M,file);apiRegister(c0d,files);$(document)[h8a](w8L,function(e,ctx,json){var H8B='dt';var p9F="space";var Y9p=f9Cm$[228782];Y9p+=K18;Y9p+=Q6Q;Y9p+=B12;var E3v=h2d;E3v+=p9F;if(e[E3v] !== H8B){return;}if(json && json[Y9p]){$[a8N](json[q2Y],function(name,filesIn){if(!Editor[q2Y][name]){Editor[q2Y][name]={};}$[d7H](Editor[q2Y][name],filesIn);});}});var _buttons=$[f9Cm$.E4X][f9Cm$.m96][O1R][j0h];$[O0B](_buttons,{create:{action:function(e,dt,node,config){var k2W="sage";var m9e="rmOpt";var n$R="proces";var l1i="rmT";var q7s="formMes";var B2v=f9Cm$[228782];B2v+=f9Cm$[23424];B2v+=m9e;B2v+=g4D;var o0h=f9Cm$.J4L;o0h+=g0r;o0h+=f9Cm$.t_T;var W2b=q4o;W2b+=Q_A;W2b+=f9Cm$.t_T;var l2g=K18;l2g+=W8E;l2g+=t1_;var I5D=g2v;I5D+=l1i;I5D+=K18;I5D+=D3E;var O5D=W1R;O5D+=M6l;var b8O=q9Z;b8O+=O4Y;b8O+=k6$;b8O+=f9Cm$.t_T;var c1X=K18;c1X+=W8E;c1X+=E49;c1X+=f9Cm$[481343];var v6L=q7s;v6L+=k2W;var W3v=X2q;W3v+=N3C;var h$O=f9Cm$[23424];h$O+=L$V;var y6H=n$R;y6H+=f9Cm$.J3V;y6H+=B96;y6H+=h7H;var that=this;var editor=config[Z9K];this[y6H](X17);editor[h$O](d_M,function(){var y8e=G2J;y8e+=x_L;y8e+=B96;I8Z.m_d();y8e+=h7H;that[y8e](h4R);})[G75]($[W3v]({buttons:config[S$p],message:config[v6L] || editor[c1X][b8O][O5D],nest:X17,title:config[I5D] || editor[l2g][W2b][o0h]},config[B2v]));},className:F4m,editor:B3c,formButtons:{action:function(e){var p8w=k18;p8w+=S_t;this[p8w]();},text:function(editor){var C31=C9K;C31+=g8Y;var L2w=q4o;L2w+=f9Cm$.t_T;L2w+=o2e;return editor[B5C][L2w][C31];}},formMessage:B3c,formOptions:{},formTitle:B3c,text:function(dt,node,config){var E8F="ttons.c";var s0h=E2j;s0h+=L_M;var W0X=q9Z;W0X+=g7G;W0X+=U58;var v9S=J0n;I8Z.m_d();v9S+=f9Cm$.Y87;v9S+=E8F;v9S+=R2g;return dt[B5C](v9S,config[Z9K][B5C][W0X][s0h]);}},createInline:{action:function(e,dt,node,config){I8Z.j$H();var C7b="position";config[Z9K][W6L](config[C7b],config[U4K]);},className:V5e,editor:B3c,formButtons:{action:function(e){this[n3l]();},text:function(editor){var N0d=f9Cm$.J3V;N0d+=T8h;I8Z.m_d();var a9H=q9Z;a9H+=R2g;return editor[B5C][a9H][N0d];}},formOptions:{},position:H29,text:function(dt,node,config){var R$G="ns.creat";var P5D=E2j;P5D+=L_M;var t2r=q9Z;I8Z.m_d();t2r+=f9Cm$.l60;t2r+=q4u;t2r+=U58;var d$0=K18;d$0+=W8E;d$0+=E49;d$0+=f9Cm$[481343];var X3n=a3D;X3n+=R$G;X3n+=f9Cm$.t_T;var R8$=K18;R8$+=W8E;R8$+=E49;R8$+=f9Cm$[481343];return dt[R8$](X3n,config[Z9K][d$0][t2r][P5D]);}},edit:{action:function(e,dt,node,config){var O_A="formBu";var R92="essa";var U6h="ormOptions";var O88="ormM";var H0v="xes";var D2d=f9Cm$[228782];D2d+=U6h;var h1P=f9Cm$.J4L;h1P+=g8Y;h1P+=k_d;var S0q=f9Cm$.t_T;S0q+=f9Cm$[555616];S0q+=K18;S0q+=f9Cm$.J4L;var d9r=K18;d9r+=W8E;d9r+=t1_;var o7k=D6p;o7k+=R92;o7k+=h7H;o7k+=f9Cm$.t_T;var z_p=f9Cm$.t_T;z_p+=D9$;var W61=f9Cm$[228782];W61+=O88;W61+=f9Cm$.t_T;W61+=M6l;var y_F=O_A;y_F+=f9Cm$.J4L;y_F+=L_M;y_F+=f9Cm$.J3V;var U66=l1f;U66+=T52;var B9c=G3n;B9c+=f9Cm$.J4L;var M2t=B96;M2t+=R3S;M2t+=H0v;var d4R=a_T;d4R+=D80;var that=this;var editor=config[Z9K];var rows=dt[d4R]({selected:X17})[M2t]();var columns=dt[m$k]({selected:X17})[f_s]();var cells=dt[q5W]({selected:X17})[f_s]();var items=columns[B97] || cells[B97]?{cells:cells,columns:columns,rows:rows}:rows;this[T0W](X17);editor[X9O](d_M,function(){var u32=B1d;u32+=f9Cm$.l60;u32+=J_2;I8Z.m_d();u32+=s29;that[u32](h4R);})[B9c](items,$[U66]({buttons:config[y_F],message:config[W61] || editor[B5C][z_p][o7k],nest:X17,title:config[X1Y] || editor[d9r][S0q][h1P]},config[D2d]));},className:g0T,editor:B3c,extend:y4I,formButtons:{action:function(e){I8Z.m_d();this[n3l]();},text:function(editor){var c2n=f9Cm$.t_T;c2n+=f9Cm$[555616];c2n+=g8Y;return editor[B5C][c2n][n3l];}},formMessage:B3c,formOptions:{},formTitle:B3c,text:function(dt,node,config){var O_T='buttons.edit';var b4V=E2j;b4V+=L_M;var T_V=h__;T_V+=E49;T_V+=f9Cm$[481343];var O$M=f9Cm$.t_T;O$M+=h44;O$M+=f9Cm$.J4L;O$M+=A5h;return dt[B5C](O_T,config[O$M][T_V][C6T][b4V]);}},remove:{action:function(e,dt,node,config){var v0W="formMessage";var g5T="formOp";var v8v="processin";var A$6="preOp";var l3o=g5T;l3o+=Y$F;var t$v=f9Cm$.J4L;t$v+=K18;t$v+=p0O;t$v+=f9Cm$.t_T;var j$9=R27;j$9+=f9Cm$.t_T;j$9+=f9Cm$.J3V;var j5S=f9Cm$.l60;j5S+=Y3n;var m79=A$6;m79+=n3_;var U5y=v8v;U5y+=h7H;var v$p=G3n;v$p+=H$0;v$p+=f9Cm$.l60;var that=this;var editor=config[v$p];this[U5y](X17);editor[X9O](m79,function(){var V3A="process";var d$j=V3A;d$j+=s29;that[d$j](h4R);})[j5S](dt[X9o]({selected:X17})[j$9](),$[d7H]({buttons:config[S$p],message:config[v0W],nest:X17,title:config[X1Y] || editor[B5C][h6J][t$v]},config[l3o]));},className:W6p,editor:B3c,extend:Q0Y,formButtons:{action:function(e){var q$J="bmit";var q8l=N5j;q8l+=q$J;I8Z.j$H();this[q8l]();},text:function(editor){var g55=f9Cm$.J3V;g55+=f9Cm$.Y87;I8Z.j$H();g55+=J0n;g55+=S_t;return editor[B5C][h6J][g55];}},formMessage:function(editor,dt){var P7D="irm";var Q7U="nfi";var d65="confirm";var g$S=D3W;g$S+=h5h;var G1N=A8G;G1N+=Q6Q;G1N+=f9Cm$.e08;G1N+=z8x;I8Z.j$H();var D0C=A6T;D0C+=Q7U;D0C+=b1Z;var K8E=k_d;K8E+=f9Cm$[481343];K8E+=h7H;K8E+=K5v;var u$N=f9Cm$.J3V;u$N+=f9Cm$.J4L;u$N+=o46;u$N+=L47;var P0v=E57;P0v+=f9Cm$[228782];P0v+=P7D;var h65=f9Cm$.l60;h65+=f9Cm$.t_T;h65+=u_A;h65+=a6R;var rows=dt[X9o]({selected:X17})[f_s]();var i18n=editor[B5C][h65];var question=typeof i18n[P0v] === u$N?i18n[d65]:i18n[d65][rows[B97]]?i18n[d65][rows[K8E]]:i18n[D0C][j7g];return question[G1N](/%d/g,rows[g$S]);},formOptions:{},formTitle:B3c,limitTo:[x8c],text:function(dt,node,config){var h0Z='buttons.remove';var Q_m=J0n;Q_m+=U2j;Q_m+=f9Cm$[23424];Q_m+=f9Cm$[481343];var Q$P=G3n;Q$P+=L2U;return dt[B5C](h0Z,config[Q$P][B5C][h6J][Q_m]);}}});_buttons[L2Q]=$[d7H]({},_buttons[a6v]);_buttons[l11][J5Q]=q$Z;_buttons[N5e]=$[d7H]({},_buttons[Q26]);_buttons[N5e][u2e]=P2p;if(!DataTable || !DataTable[d4l] || !DataTable[Q5K](J5M)){throw new Error(f8u);}var Editor=(function(){var x9v="models";var Z2a="factory";var V_0="internalMultiInfo";var V3k="ternalSett";var p1I="ernalEvent";var L7n="defa";var b5Y='2.1.3';var p3V="ults";var T9K="version";var N0L="Sources";var d8_="int";var p3_="internalI18n";var l7I="uploa";var r$m=s8o;r$m+=n7x;var t4x=f9Cm$[555616];t4x+=q94;t4x+=N0L;var V4Z=L7n;V4Z+=p3V;var V2s=l7I;V2s+=f9Cm$[555616];var R8S=B1d;R8S+=i0g;R8S+=f9Cm$.l60;R8S+=f9Cm$.J3V;var b3w=f07;b3w+=i3q;b3w+=f9Cm$.t_T;b3w+=f9Cm$.J3V;var S83=B96;S83+=V3k;S83+=s29;S83+=f9Cm$.J3V;var S7g=G2J;S7g+=V9Y;var V3z=d8_;V3z+=p1I;var e47=u6g;e47+=O6L;function Editor(init,cjsJq){var T44='"></div></div>';var y4G="nN";var S_e='<div data-dte-e="head" class="';var s0K="displayNode";var I$$="messa";var M4o="_fieldF";var q3K="und";var d5y="unique";var n64="den";var r_Q="tent";var I5T=" cl";var I_X="\"></";var Y07="dicator";var U3F="orm_con";var y_9="topen";var p8e="layReorder";var f5a="iqu";var b8w="disable";var T03="inE";var j$6="_mult";var J8U='<div data-dte-e="body" class="';var W$s='init.dt.dte';var f9l="_pos";var u3o='<div data-dte-e="form_content" class="';var A6W="depen";var b$g="gro";var x_8="efaul";var I9A="iv data-dte-e=\"form_buttons\" class";var J3$="nlin";var f3M="els";var Z3r="back";var o2b="Nod";var H3R='"><div class="';var W87="DataTables Editor must be initialised as a \'new\' ins";var u77="domTable";var u6q="oy";var J42="destr";var t8e="lePosition";var l0u="i18n.";var P2V="=\"";var V2$="ttin";var v0A="_submitT";var I$c=".dte";var p0e="crudAr";var s8x='initComplete';var c04="\"foot\" class=\"";var z3d="_submitSuc";var u0X='<div data-dte-e="body_content" class="';var G0F="y_";var q1M='<div data-dte-e="form_info" class="';var x4m="rro";var W2n='<div data-dte-e="form_error" class="';var M5o="i18";var t_r="_nestedClose";var e8w="ubble";var n_J='<form data-dte-e="form" class="';var y4E='foot';var O5L="<div data-dte-e=\"processing\"";var c0_="orde";var U9m='Cannot find display controller ';var W2V="init";var b6b="embleMai";var t13="rom";var m3M="tance";var k8n="fac";var C7D="rap";var C0E="mic";var Z0u="_clearDyna";var u7k="xhr.d";var H5q="<div data-dte-e=";var H_v="lear";var q6z="_optionsUpdate";var V5f=B96;V5f+=g8Y;V5f+=i2S;V5f+=W9u;var D2q=f9Cm$[555616];D2q+=E7j;D2q+=I8$;var a_9=f9Cm$[555616];a_9+=K18;a_9+=V1c;a_9+=I8$;var M6P=C1Z;M6P+=b5Q;M6P+=g5f;var C2v=f9Cm$[555616];C2v+=K18;C2v+=C3S;C2v+=g5f;var n36=u7k;n36+=f9Cm$.J4L;n36+=I$c;var K51=l0u;K51+=F3v;K51+=I$c;var W7e=f9Cm$.Y87;W7e+=f9Cm$[481343];W7e+=f5a;W7e+=f9Cm$.t_T;var G$N=s4L;G$N+=o$D;var Q1D=f9Cm$[555616];Q1D+=f9Cm$[23424];Q1D+=D6p;var p$d=n_A;p$d+=f9Cm$.J3V;var D8T=y9e;D8T+=y1r;var F19=f9Cm$[228782];F19+=A5h;F19+=D6p;var m4p=f9Cm$.t_T;m4p+=f9Cm$.l60;m4p+=f9Cm$.l60;m4p+=A5h;var w9N=f9Cm$[228782];w9N+=U3F;w9N+=U58;w9N+=l1E;var E91=u_6;E91+=k8A;E91+=T5B;var a2n=f9Cm$[228782];a2n+=f9Cm$[23424];a2n+=f9Cm$.l60;a2n+=D6p;var l7Z=q7L;l7Z+=I9A;l7Z+=v19;l7Z+=y9e;var I30=r0K;I30+=f9Cm$[555616];I30+=G0F;I30+=P8l;var K6W=f9Cm$[555616];K6W+=f9Cm$[23424];K6W+=D6p;var N02=Z4$;N02+=O94;N02+=i$Z;var v3c=f9Cm$[228782];v3c+=f9Cm$[23424];v3c+=f9Cm$.l60;v3c+=D6p;var O$4=f9Cm$[228782];O$4+=f9Cm$[23424];O$4+=f9Cm$.l60;O$4+=D6p;var n9j=o0P;n9j+=x2W;var o6Z=I_X;o6Z+=h8B;o6Z+=i$Z;var J0M=q9Z;J0M+=h8a;J0M+=r_Q;var M7p=y9e;M7p+=i$Z;var U3v=H5q;U3v+=c04;var H6n=J0n;H6n+=f9Cm$[23424];H6n+=f9Cm$[555616];H6n+=g5f;var E25=B96;E25+=Y07;var j2S=B1d;j2S+=f9Cm$.l60;j2S+=J_2;j2S+=s29;var d$V=O5L;d$V+=I5T;d$V+=S$6;d$V+=P2V;var l3O=H3b;l3O+=C7D;l3O+=B1d;l3O+=F7I;var L9F=Y1m;L9F+=V2$;L9F+=M8F;var a6A=L$j;a6A+=f3M;var Z7v=K18;Z7v+=W8E;Z7v+=E49;Z7v+=f9Cm$[481343];var V6E=M5o;V6E+=f9Cm$[481343];var a6N=r8u;a6N+=P_W;var H_M=f9Cm$.J4L;H_M+=H3O;H_M+=f9Cm$.t_T;var V9r=K18;V9r+=r4b;V9r+=f9Cm$.l60;V9r+=q9Z;var g8c=u6E;g8c+=n59;g8c+=y4G;g8c+=o1Y;var a2m=L$j;a2m+=f3M;I8Z.j$H();var h8Y=X2q;h8Y+=f9Cm$.J4L;h8Y+=n3_;h8Y+=f9Cm$[555616];var N8g=f9Cm$[555616];N8g+=x_8;N8g+=f9Cm$.E9M;var w5G=k8n;w5G+=H$0;w5G+=u8z;var P3P=g4c;P3P+=d0b;var j$k=z3d;j$k+=x_L;var f$g=v0A;f$g+=l27;var P_w=J04;P_w+=f9Cm$.Y87;P_w+=f2C;P_w+=f9Cm$.J4L;var V2x=f9l;V2x+=y_9;var t4P=j$6;t4P+=K18;t4P+=h3G;t4P+=q9g;var E8h=j7g;E8h+=I$$;E8h+=P9U;var i_F=z7h;i_F+=J3$;i_F+=f9Cm$.t_T;var Q$q=v7I;Q$q+=s4W;Q$q+=L_b;Q$q+=B12;var d4a=M4o;d4a+=t13;d4a+=o2b;d4a+=f9Cm$.t_T;var q6c=p3g;q6c+=f9Cm$[555616];q6c+=g8Y;var d1g=I_j;d1g+=E7j;d1g+=p8e;var W5$=j7g;W5$+=p0e;W5$+=h7H;W5$+=f9Cm$.J3V;var R0h=j7g;R0h+=q9Z;R0h+=C2Q;R0h+=f9Cm$.t_T;var v6P=Z0u;v6P+=C0E;v6P+=Y5K;v6P+=f9Cm$[23424];var F0N=j7g;F0N+=J0n;F0N+=c5p;F0N+=f9Cm$.l60;var O7I=w57;O7I+=b6b;O7I+=f9Cm$[481343];var x7h=O1r;x7h+=f9Cm$.Z$r;x7h+=J_Y;var g8w=O1r;g8w+=c7u;g8w+=s_b;g8w+=S81;var x0p=I7r;x0p+=f9Cm$.e08;x0p+=Q6Q;var p4c=k1a;p4c+=f9Cm$.J4L;var r_I=f9Cm$.J3V;r_I+=s$2;r_I+=h98;var a7p=c0_;a7p+=f9Cm$.l60;var p5V=W1R;p5V+=M6l;var M_Q=T03;M_Q+=x4m;M_Q+=f9Cm$.l60;var V3e=K18;V3e+=f9Cm$[555616];V3e+=f9Cm$.J3V;var W9b=P9U;W9b+=f9Cm$.J4L;var U20=f9Cm$[228782];U20+=K18;U20+=Q6Q;U20+=f9Cm$.t_T;var o0V=w3$;o0V+=g8Y;var t6f=J42;t6f+=u6q;var D3T=A6W;D3T+=n64;D3T+=f9Cm$.J4L;var V4Y=q9Z;V4Y+=H_v;var t50=q1i;t50+=I4O;t50+=W4z;var R_e=C_B;R_e+=t8e;var R4J=J0n;R4J+=e8w;var L4R=R6r;L4R+=f9Cm$.Y87;L4R+=f9Cm$.l60;var c7v=Z3r;c7v+=b$g;c7v+=q3K;var j7X=f9Cm$.e08;j7X+=f9Cm$[555616];j7X+=f9Cm$[555616];var _this=this;this[j7X]=add;this[N7l]=ajax;this[c7v]=background;this[L4R]=blur;this[R4J]=bubble;this[R_e]=bubblePosition;this[t50]=buttons;this[V4Y]=clear;this[H9B]=close;this[G75]=create;this[p2n]=undependent;this[D3T]=dependent;this[t6f]=destroy;this[b8w]=disable;this[G7z]=display;this[n5R]=displayed;this[s0K]=displayNode;this[o0V]=edit;this[y9q]=enable;this[h8G]=error$1;this[D_P]=field;this[P5P]=fields;this[U20]=file;this[q2Y]=files;this[W9b]=get;this[J7j]=hide;this[V3e]=ids;this[M_Q]=inError;this[n0T]=inline;this[W6L]=inlineCreate;this[p5V]=message;this[h56]=mode;this[n9R]=modifier;this[o1D]=multiGet;this[Y$8]=multiSet;this[z$j]=node;this[v93]=off;this[h8a]=on;this[X9O]=one;this[N4D]=open;this[a7p]=order;this[h6J]=remove;this[D0u]=set;this[r_I]=show;this[p4c]=submit;this[z6u]=table;this[b2T]=template;this[M3D]=title;this[x0p]=val;this[g8w]=_actionClass;this[x7h]=_ajax;this[T4f]=_animate;this[O7I]=_assembleMain;this[F0N]=_blur;this[v6P]=_clearDynamicInfo;this[R0h]=_close;this[k7m]=_closeReg;this[W5$]=_crudArgs;this[e3V]=_dataSource;this[d1g]=_displayReorder;this[q6c]=_edit;this[G29]=_event;this[A1x]=_eventName;this[d4a]=_fieldFromNode;this[Q$q]=_fieldNames;this[U5L]=_focus;this[Q49]=_formOptions;this[i_F]=_inline;this[I45]=_inputTrigger;this[q6z]=_optionsUpdate;this[E8h]=_message;this[t4P]=_multiInfo;this[t_r]=_nestedClose;this[Y1l]=_nestedOpen;this[V2x]=_postopen;this[w6o]=_preopen;this[I0J]=_processing;this[S3a]=_noProcessing;this[P_w]=_submit;this[f$g]=_submitTable;this[j$k]=_submitSuccess;this[C4S]=_submitError;this[P3P]=_tidy;this[f5n]=_weakInArray;if(Editor[w5G](init,cjsJq)){return Editor;}if(!(this instanceof Editor)){var M2p=W87;M2p+=m3M;alert(M2p);}init=$[d7H](X17,{},Editor[N8g],init);this[q9Z]=init;this[f9Cm$.J3V]=$[h8Y](X17,{},Editor[a2m][q$G],{actionName:init[g8c],ajax:init[N7l],formOptions:init[U4K],idSrc:init[V9r],table:init[u77] || init[H_M],template:init[b2T]?$(init[b2T])[Z3_]():B3c});this[A2x]=$[d7H](X17,{},Editor[a6N]);this[V6E]=init[Z7v];Editor[a6A][L9F][d5y]++;var that=this;var classes=this[A2x];var wrapper=$(v4Q + classes[l3O] + x$l + d$V + classes[j2S][E25] + R4f + J8U + classes[f9H][A2A] + x$l + u0X + classes[H6n][P8l] + g42 + Q4m + U3v + classes[J$g][A2A] + M7p + v4Q + classes[J$g][J0M] + o6Z + Q4m + n9j);var form=$(n_J + classes[O$4][R$f] + x$l + u3o + classes[v3c][P8l] + g42 + N02);this[K6W]={body:el(p9T,wrapper)[C37],bodyContent:el(I30,wrapper)[C37],buttons:$(l7Z + classes[a2n][d7J] + E91)[C37],footer:el(y4E,wrapper)[C37],form:form[C37],formContent:el(w9N,form)[C37],formError:$(W2n + classes[O94][m4p] + g42)[C37],formInfo:$(q1M + classes[F19][c0T] + D8T)[C37],header:$(S_e + classes[L2d][A2A] + H3R + classes[L2d][P8l] + T44)[C37],processing:el(H27,wrapper)[C37],wrapper:wrapper[C37]};$[a8N](init[p$d],function(evt,fn){var Q1w=f9Cm$[23424];I8Z.j$H();Q1w+=f9Cm$[481343];that[Q1w](evt,function(){I8Z.m_d();var T2v="ply";var Z$F=f9Cm$.e08;Z$F+=B1d;Z$F+=T2v;var argsIn=[];for(var _i=C37;_i < arguments[B97];_i++){argsIn[_i]=arguments[_i];}fn[Z$F](that,argsIn);});});this[Q1D];if(init[G$N]){this[g66](init[P5P]);}$(document)[h8a](W$s + this[f9Cm$.J3V][W7e],function(e,settings,json){var h0e="Ap";var table=_this[f9Cm$.J3V][z6u];if(table){var y6A=f9Cm$[481343];y6A+=f9Cm$[23424];y6A+=f9Cm$[555616];y6A+=f9Cm$.t_T;var b1z=f9Cm$.J4L;b1z+=f9Cm$.e08;b1z+=V5s;var x3X=f9Cm$[481343];x3X+=u08;x3X+=l27;var T0K=h0e;T0K+=K18;var dtApi=new DataTable[T0K](table);if(settings[x3X] === dtApi[b1z]()[y6A]()){var m01=p8d;m01+=N7v;settings[m01]=_this;}}})[h8a](K51 + this[f9Cm$.J3V][d5y],function(e,settings){I8Z.j$H();var Q_D="oLanguage";var o7e="nTable";var G1O="ngu";var table=_this[f9Cm$.J3V][z6u];if(table){var M9m=f9Cm$[481343];M9m+=q9c;M9m+=f9Cm$.t_T;var e8y=z5S;e8y+=B1d;e8y+=K18;var dtApi=new DataTable[e8y](table);if(settings[o7e] === dtApi[z6u]()[M9m]()){var m9T=f9Cm$[23424];m9T+=D8m;m9T+=G1O;m9T+=D2Y;if(settings[m9T][Z9K]){var Q_Q=l1f;Q_Q+=f9Cm$[481343];Q_Q+=f9Cm$[555616];$[Q_Q](X17,_this[B5C],settings[Q_D][Z9K]);}}}})[h8a](n36 + this[f9Cm$.J3V][d5y],function(e,settings,json){I8Z.j$H();var d9s="Tabl";var d9c=Q3e;d9c+=V5s;var table=_this[f9Cm$.J3V][d9c];if(table){var e1Q=f9Cm$[481343];e1Q+=d9s;e1Q+=f9Cm$.t_T;var W5Y=z5S;W5Y+=B1d;W5Y+=K18;var dtApi=new DataTable[W5Y](table);if(settings[e1Q] === dtApi[z6u]()[z$j]()){_this[q6z](json);}}});if(!Editor[C2v][init[M6P]]){throw new Error(U9m + init[G7z]);}this[f9Cm$.J3V][I2G]=Editor[a_9][init[D2q]][W2V](this);this[G29](s8x,[]);$(document)[c6V](V5f,[this]);}Editor[e47][V3z]=function(name,args){I8Z.m_d();this[G29](name,args);};Editor[P99][p3_]=function(){var x4E=h__;x4E+=E49;I8Z.j$H();x4E+=f9Cm$[481343];return this[x4E];};Editor[P99][V_0]=function(){return this[m_q]();};Editor[S7g][S83]=function(){I8Z.m_d();return this[f9Cm$.J3V];};Editor[t3_]={checkbox:checkbox,datatable:datatable,datetime:datetime,hidden:hidden,password:password,radio:radio,readonly:readonly,select:select,text:text,textarea:textarea,upload:upload,uploadMany:uploadMany};Editor[q2Y]={};Editor[T9K]=b5Y;Editor[b3w]=classNames;Editor[x6i]=Field;Editor[i7B]=B3c;Editor[h8G]=error;Editor[R8S]=pairs;Editor[Z2a]=factory;Editor[V2s]=upload$1;Editor[V4Z]=defaults$1;Editor[x9v]={button:button,displayController:displayController,fieldType:fieldType,formOptions:formOptions,settings:settings};Editor[t4x]={dataTable:dataSource$1,html:dataSource};I8Z.j$H();Editor[r$m]={envelope:envelope,lightbox:self};Editor[L2y]=function(id){I8Z.j$H();return safeDomId(id,j_l);};return Editor;})();DataTable[h7n]=Editor;$[B2l][Y7b][G_n]=Editor;if(DataTable[O0c]){Editor[i7B]=DataTable[i7B];}if(DataTable[H_L][i$O]){var e7a=T_3;e7a+=X5O;var T5n=f9Cm$.t_T;T5n+=M27;var X7y=i9j;X7y+=f9Cm$[555616];$[X7y](Editor[t3_],DataTable[T5n][e7a]);}DataTable[H_L][B4g]=Editor[D4U];return Editor;});})();

/*! jQuery UI integration for DataTables' Editor
 * © SpryMedia Ltd - datatables.net/license
 */

(function( factory ){
	if ( typeof define === 'function' && define.amd ) {
		// AMD
		define( ['jquery', 'datatables.net-jqui', 'datatables.net-editor'], function ( $ ) {
			return factory( $, window, document );
		} );
	}
	else if ( typeof exports === 'object' ) {
		// CommonJS
		var jq = require('jquery');
		var cjsRequires = function (root, $) {
			if ( ! $.fn.dataTable ) {
				require('datatables.net-jqui')(root, $);
			}

			if ( ! $.fn.dataTable.Editor ) {
				require('datatables.net-editor')(root, $);
			}
		};

		if (typeof window === 'undefined') {
			module.exports = function (root, $) {
				if ( ! root ) {
					// CommonJS environments without a window global must pass a
					// root. This will give an error otherwise
					root = window;
				}

				if ( ! $ ) {
					$ = jq( root );
				}

				cjsRequires( root, $ );
				return factory( $, root, root.document );
			};
		}
		else {
			cjsRequires( window, jq );
			module.exports = factory( jq, window, window.document );
		}
	}
	else {
		// Browser
		factory( jQuery, window, document );
	}
}(function( $, window, document, undefined ) {
'use strict';
var DataTable = $.fn.dataTable;



var Editor = DataTable.Editor;
var doingClose = false;

/*
 * Set the default display controller to be our foundation control 
 */
Editor.defaults.display = "jqueryui";

/*
 * Change the default classes from Editor to be classes for Bootstrap
 */
var buttonClass = "btn ui-button ui-widget ui-state-default ui-corner-all ui-button-text-only";
$.extend( true, $.fn.dataTable.Editor.classes, {
	form: {
		button:  buttonClass,
		buttonInternal:  buttonClass
	}
} );

var dialouge;
var shown = false;

/*
 * jQuery UI display controller - this is effectively a proxy to the jQuery UI
 * modal control.
 */
Editor.display.jqueryui = $.extend( true, {}, Editor.models.displayController, {
	init: function ( dte ) {
		if (! dialouge) {
			dialouge = $('<div class="DTED"></div>')
				.css('display', 'none')
				.appendTo('body')
				.dialog( $.extend( true, Editor.display.jqueryui.modalOptions, {
					autoOpen: false,
					buttons: { "A": function () {} }, // fake button so the button container is created
					closeOnEscape: false // allow editor's escape function to run
				} ) );
		}

		return Editor.display.jqueryui;
	},

	open: function ( dte, append, callback ) {
		dialouge
			.children()
			.detach();

		dialouge
			.append( append )
			.dialog( 'open' );

		$(dte.dom.formError).appendTo(
			dialouge.parent().find('div.ui-dialog-buttonpane')
		);

		dialouge.parent().find('.ui-dialog-title').html( dte.dom.header.innerHTML );
		dialouge.parent().addClass('DTED');

		// Modify the Editor buttons to be jQuery UI suitable
		var buttons = $(dte.dom.buttons)
			.children()
			.addClass( 'ui-button ui-widget ui-state-default ui-corner-all ui-button-text-only' )
			.each( function () {
				$(this).wrapInner( '<span class="ui-button-text"></span>' );
			} );

		// Move the buttons into the jQuery UI button set
		dialouge.parent().find('div.ui-dialog-buttonset')
			.children()
			.detach();

		dialouge.parent().find('div.ui-dialog-buttonset')
			.append( buttons.parent() );

		dialouge
			.parent()
			.find('button.ui-dialog-titlebar-close')
			.off('click.dte-ju')
			.on('click.dte-ju', function () {
				dte.close('icon');
			});

		// Need to know when the dialogue is closed using its own trigger
		// so we can reset the form
		$(dialouge)
			.off( 'dialogclose.dte-ju' )
			.on( 'dialogclose.dte-ju', function (e) {
				if ( ! doingClose ) {
					dte.close();
				}
			} );

		shown = true;

		if ( callback ) {
			callback();
		}
	},

	close: function ( dte, callback ) {
		if ( dialouge ) {
			// Don't want to trigger a close() call from dialogclose!
			doingClose = true;
			dialouge.dialog( 'close' );
			doingClose = false;
		}

		shown = false;

		if ( callback ) {
			callback();
		}
	},

	node: function ( dte ) {
		return dialouge[0];
	},

	// jQuery UI dialogues perform their own focus capture
	captureFocus: false
} );


Editor.display.jqueryui.modalOptions = {
	width: 600,
	modal: true
};


return Editor;
}));

const UIDataTablelibraryLoadedEvent = new CustomEvent('UIDataTable_libraryLoaded');

document.dispatchEvent(UIDataTablelibraryLoadedEvent);
