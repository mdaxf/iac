/*
	Author: Rocky WANG
	Date: 07/30/2020
	
	The svc-gantt is only used for the Service Accelerator project or other projects which authorized by the Author. All others actions, for instance, copy, reference the library are not allowed. 
*/

var Gantt = (function () {
'use strict';

const YEAR = 'year';
const MONTH = 'month';
const DAY = 'day';
const HOUR = 'hour';
const MINUTE = 'minute';
const SECOND = 'second';
const MILLISECOND = 'millisecond';

const month_names = {
    en: [
        'January',
        'February',
        'March',
        'April',
        'May',
        'June',
        'July',
        'August',
        'September',
        'October',
        'November',
        'December'
    ],
    jp: [
        '1月',
        '2月',
        '3月',
        '4月',
        '5月',
        '6月',
        '7月',
        '8月',
        '9月',
        '10月',
        '11月',
        '12月'
    ],
    ru: [
        'Январь',
        'Февраль',
        'Март',
        'Апрель',
        'Май',
        'Июнь',
        'Июль',
        'Август',
        'Сентябрь',
        'Октябрь',
        'Ноябрь',
        'Декабрь'
    ]
};

var date_utils = {
    parse(date, date_separator = '-', time_separator = /[.:]/) {
        if (date instanceof Date) {
            return date;
        }
        if (typeof date === 'string') {
            let date_parts, time_parts;
            //const parts = date.split(' ');
			const parts = date.split('T');
            date_parts = parts[0]
                .split(date_separator)
                .map(val => parseInt(val, 10));
            time_parts = parts[1] && parts[1].split(time_separator);

            // month is 0 indexed
            date_parts[1] = date_parts[1] - 1;

            let vals = date_parts;

            if (time_parts && time_parts.length) {
                if (time_parts.length == 4) {
                    time_parts[3] = '0.' + time_parts[3];
                    time_parts[3] = parseFloat(time_parts[3]) * 1000;
                }
                vals = vals.concat(time_parts);
            }

            return new Date(...vals);
        }
    },

    to_string(date, with_time = false) {
        if (!(date instanceof Date)) {
            throw new TypeError('Invalid argument type');
        }
        const vals = this.get_date_values(date).map((val, i) => {
            if (i === 1) {
                // add 1 for month
                val = val + 1;
            }

            if (i === 6) {
                return padStart(val + '', 3, '0');
            }

            return padStart(val + '', 2, '0');
        });
        const date_string = `${vals[0]}-${vals[1]}-${vals[2]}`;
        const time_string = `${vals[3]}:${vals[4]}:${vals[5]}.${vals[6]}`;

        return date_string + (with_time ? ' ' + time_string : '');
    },

    format(date, format_string = 'YYYY-MM-DD HH:mm:ss.SSS', lang = 'en') {
        const values = this.get_date_values(date).map(d => padStart(d, 2, 0));
        const format_map = {
            YYYY: values[0],
            MM: padStart(+values[1] + 1, 2, 0),
            DD: values[2],
            HH: values[3],
            mm: values[4],
            ss: values[5],
            SSS:values[6],
            D: values[2],
            MMMM: month_names[lang][+values[1]],
            MMM: month_names[lang][+values[1]]
        };

        let str = format_string;
        const formatted_values = [];

        Object.keys(format_map)
            .sort((a, b) => b.length - a.length) // big string first
            .forEach(key => {
                if (str.includes(key)) {
                    str = str.replace(key, `$${formatted_values.length}`);
                    formatted_values.push(format_map[key]);
                }
            });

        formatted_values.forEach((value, i) => {
            str = str.replace(`$${i}`, value);
        });

        return str;
    },

    diff(date_a, date_b, scale = DAY) {
        let milliseconds, seconds, hours, minutes, days, months, years;

        milliseconds = date_a - date_b;
        seconds = milliseconds / 1000;
        minutes = seconds / 60;
        hours = minutes / 60;
        days = hours / 24;
        months = days / 30;
        years = months / 12;

        if (!scale.endsWith('s')) {
            scale += 's';
        }

        return Math.floor(
            {
                milliseconds,
                seconds,
                minutes,
                hours,
                days,
                months,
                years
            }[scale]
        );
    },

	
    today() {
        const vals = this.get_date_values(new Date()).slice(0, 3);
        return new Date(...vals);
    },

    now() {
        return new Date();
    },

    add(date, qty, scale) {
        qty = parseInt(qty, 10);
        const vals = [
            date.getFullYear() + (scale === YEAR ? qty : 0),
            date.getMonth() + (scale === MONTH ? qty : 0),
            date.getDate() + (scale === DAY ? qty : 0),
            date.getHours() + (scale === HOUR ? qty : 0),
            date.getMinutes() + (scale === MINUTE ? qty : 0),
            date.getSeconds() + (scale === SECOND ? qty : 0),
            date.getMilliseconds() + (scale === MILLISECOND ? qty : 0)
        ];
        return new Date(...vals);
    },

    start_of(date, scale) {
        const scores = {
            [YEAR]: 6,
            [MONTH]: 5,
            [DAY]: 4,
            [HOUR]: 3,
            [MINUTE]: 2,
            [SECOND]: 1,
            [MILLISECOND]: 0
        };

        function should_reset(_scale) {
            const max_score = scores[scale];
            return scores[_scale] <= max_score;
        }

        const vals = [
            date.getFullYear(),
            should_reset(YEAR) ? 0 : date.getMonth(),
            should_reset(MONTH) ? 1 : date.getDate(),
            should_reset(DAY) ? 0 : date.getHours(),
            should_reset(HOUR) ? 0 : date.getMinutes(),
            should_reset(MINUTE) ? 0 : date.getSeconds(),
            should_reset(SECOND) ? 0 : date.getMilliseconds()
        ];

        return new Date(...vals);
    },

    clone(date) {
        return new Date(...this.get_date_values(date));
    },

    get_date_values(date) {
        return [
            date.getFullYear(),
            date.getMonth(),
            date.getDate(),
            date.getHours(),
            date.getMinutes(),
            date.getSeconds(),
            date.getMilliseconds()
        ];
    },

    get_days_in_month(date) {
        const no_of_days = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];

        const month = date.getMonth();

        if (month !== 1) {
            return no_of_days[month];
        }

        // Feb
        const year = date.getFullYear();
        if ((year % 4 == 0 && year % 100 != 0) || year % 400 == 0) {
            return 29;
        }
        return 28;
    }
};

// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/String/padStart
function padStart(str, targetLength, padString) {
    str = str + '';
    targetLength = targetLength >> 0;
    padString = String(typeof padString !== 'undefined' ? padString : ' ');
    if (str.length > targetLength) {
        return String(str);
    } else {
        targetLength = targetLength - str.length;
        if (targetLength > padString.length) {
            padString += padString.repeat(targetLength / padString.length);
        }
        return padString.slice(0, targetLength) + String(str);
    }
}

function $(expr, con) {
    return typeof expr === 'string'
        ? (con || document).querySelector(expr)
        : expr || null;
}

async function asycreateSVG(tag, attrs){
	//window.requestAnimationFrame(() => createSVG(tag, attrs));
	return createSVG(tag, attrs);
	
}

function createSVG(tag, attrs) {
    const elem = document.createElementNS('http://www.w3.org/2000/svg', tag);
	
	if(elem == null || !elem)
		return null;
	
    for (let attr in attrs) {
        if (attr === 'append_to') {
            const parent = attrs.append_to;
            parent.appendChild(elem);
        } else if (attr === 'innerHTML') {
            elem.innerHTML = attrs.innerHTML;
        } else if(attr === 'fontsize'){
			elem.setAttribute('font-size', attrs[attr]);
		}
		else {
            elem.setAttribute(attr, attrs[attr]);
        }
    }
    return elem;
}

function animateSVG(svgElement, attr, from, to) {
    const animatedSvgElement = getAnimationElement(svgElement, attr, from, to);

    if (animatedSvgElement === svgElement) {
        // triggered 2nd time programmatically
        // trigger artificial click event
        const event = document.createEvent('HTMLEvents');
        event.initEvent('click', true, true);
        event.eventName = 'click';
        animatedSvgElement.dispatchEvent(event);
    }
}

function getAnimationElement(
    svgElement,
    attr,
    from,
    to,
    dur = '0.4s',
    begin = '0.1s'
) {
    const animEl = svgElement.querySelector('animate');
    if (animEl) {
        $.attr(animEl, {
            attributeName: attr,
            from,
            to,
            dur,
            begin: 'click + ' + begin // artificial click
        });
        return svgElement;
    }

    const animateElement = createSVG('animate', {
        attributeName: attr,
        from,
        to,
        dur,
        begin,
        calcMode: 'spline',
        values: from + ';' + to,
        keyTimes: '0; 1',
        keySplines: cubic_bezier('ease-out')
    });
    svgElement.appendChild(animateElement);

    return svgElement;
}

function cubic_bezier(name) {
    return {
        ease: '.25 .1 .25 1',
        linear: '0 0 1 1',
        'ease-in': '.42 0 1 1',
        'ease-out': '0 0 .58 1',
        'ease-in-out': '.42 0 .58 1'
    }[name];
}

function Checkandupdatethechange(e){
	
	
}

function hidepopup(){
	$('.gantt-container').find('.popup-wrapper').hide();
}

$.on = (element, event, selector, callback) => {
	if(!element || element == null)
		return;
    if (!callback) {
        callback = selector;
        $.bind(element, event, callback);
    } else {
        $.delegate(element, event, selector, callback);
    }
};

$.off = (element, event, handler) => {
    element.removeEventListener(event, handler);
};

$.bind = (element, event, callback) => {
	if(!element || element == null)
		return;
    event.split(/\s+/).forEach(function(event) {
        element.addEventListener(event, callback);
    });
};

$.delegate = (element, event, selector, callback) => {
	if(!element || element == null)
		return;
    element.addEventListener(event, function(e) {
        const delegatedTarget = e.target.closest(selector);
        if (delegatedTarget) {
            e.delegatedTarget = delegatedTarget;
            callback.call(this, e, delegatedTarget);
        }
    });
};

$.closest = (selector, element) => {
    if (!element) return null;

    if (element.matches(selector)) {
        return element;
    }

    return $.closest(selector, element.parentNode);
};
$.clearcontent = (element) => {
	element.empty();
}
$.attr = (element, attr, value) => {
	
	if(!element || element == null)
		return;
	
    if (!value && typeof attr === 'string') {
        return element.getAttribute(attr);
    }

    if (typeof attr === 'object') {
        for (let key in attr) {
            $.attr(element, key, attr[key]);
        }
        return;
    }

    element.setAttribute(attr, value);
};

class Bar {
    constructor(gantt, task) {
        this.set_defaults(gantt, task);
        this.prepare();
        this.draw();
        this.bind();
    }

    set_defaults(gantt, task) {
        this.action_completed = false;
        this.gantt = gantt;
        this.task = task;
    }

    prepare() {
        this.prepare_values();
        this.prepare_helpers();
    }

    prepare_values() {
        this.invalid = this.task.invalid;
        this.height = this.gantt.options.bar_height;
        this.x = this.compute_x();
        this.y = this.compute_y();
        this.corner_radius = this.gantt.options.bar_corner_radius;
        this.duration =
            date_utils.diff(this.task._end, this.task._start, 'second') /
            this.gantt.options.step / 3600;
        this.width = this.gantt.options.column_width * this.duration;
		
		this.width = this.width < 10? 10: this.width; // avaoid the bar width 0 or negative
		
        this.progress_width =
            this.gantt.options.column_width *
                this.duration *
                (this.task.progress / 100) || 0;
		
		this.progress_width = this.progress_width > this.width? this.width: this.progress_width;  // avoid progress bar wider than bar 

		this.actual_bar_x = this.compute_actual_x();
		this.actual_bar_y = (this.compute_actual_y() + (this.height * 15)/100);
		
        this.group = createSVG('g', {
            class: 'bar-wrapper ' + (this.task.custom_class || ''),
            'data-id': this.task.id
        });
        this.bar_group = createSVG('g', {
            class: 'bar-group',
            append_to: this.group
        });
        this.handle_group = createSVG('g', {
            class: 'handle-group',
            append_to: this.group
        });
		
		console.log(this)
    }

    prepare_helpers() {
        SVGElement.prototype.getX = function() {
            return +this.getAttribute('x');
        };
        SVGElement.prototype.getY = function() {
            return +this.getAttribute('y');
        };
        SVGElement.prototype.getWidth = function() {
            return +this.getAttribute('width');
        };
        SVGElement.prototype.getHeight = function() {
            return +this.getAttribute('height');
        };
        SVGElement.prototype.getEndX = function() {
            return this.getX() + this.getWidth();
        };
    }

    draw() {
		if(this.x < 0 || this. y < 0)
			return;
		
        this.draw_bar();
		if(this.gantt.options.showactual){
			if(this.gantt.options.showactual == 'yes'){
				this.draw_actual_bar();
				this.draw_actual_label();
				}
		}	
        this.draw_progress_bar();
        this.draw_label();
		
		this.draw_progress_label();
        this.draw_resize_handles();
		this.draw_assigned_badges();
		
		if(this.task.issuetype && this.task.issuetype !='')
			this.draw_issue_icon(this.task.issuetype);
		else
			this.draw_issue_icon('');

    }

    draw_bar() {
        this.$bar = createSVG('rect', {
            x: this.x,
            y: this.y,
            width: this.width,
            height: this.height,
            rx: this.corner_radius,
            ry: this.corner_radius,
			orderno: this.task.orderno,
			operation:this.task.operation,
			ordertype: this.task.ordertype,
			taskid: this.task.id,
			bartype: "taskbar",
			barstatus: this.task.status,
			baroldstatus: this.task.status,
            class: 'bar ' + (!this.task.barclass) ? this.task.barclass  : '',
            append_to: this.bar_group
        });

        animateSVG(this.$bar, 'width', 0, this.width);
		
		//this.$bar.classList.add('bar');
		
        if (this.invalid) {
            this.$bar.classList.add('bar-invalid');
        }
    }

	draw_actual_bar(){
		this.actualduration =
            date_utils.diff(this.task._actualend, this.task._actualstart, 'second') /
            this.gantt.options.step / 3600;
        this.actualwidth = this.gantt.options.column_width * this.actualduration;
		
		if(this.task.progress > 0)
			this.actualwidth = this.actualwidth < 10? 10: this.actualwidth;
		
	
        this.$actualbar = createSVG('rect', {
            x: this.actual_bar_x ,// this.compute_actual_x(),
            y: this.actual_bar_y, //(this.compute_actual_y() + (this.height * 15)/100),
            width: this.actualwidth,
            height: (this.height * 70)/100,
            rx: this.corner_radius,
            ry: this.corner_radius,
            class: 'actualbar ' + !this.task.actualbarclass ? this.task.actualbarclass : '',
            append_to: this.bar_group
        });
		
		
        animateSVG(this.$actualbar, 'width', 0, this.actualwidth);

        if (this.invalid) {
            this.$actualbar.classList.add('bar-invalid');
        }		
	}
	
	remove_badge(employeeid){
		var that = this; 
		that.bar_group.querySelectorAll('circle.bar-badge').forEach(function(el){
		//	console.log((el))
			if(el.getAttribute('adid') == employeeid)
				el.remove()
		});
		that.bar_group.querySelectorAll('text.bar-badge-label').forEach(function(el){
		//	console.log((el))
			if(el.getAttribute('adid') == employeeid)
				el.remove();
		});
		
		let badges = [];
		let tempbadges = [];
		
		// badges: employee
        if (typeof this.task.badges === 'string' || !this.task.badges) {
            if (this.task.badges) {
                    badges = this.task.badges
                        .split(',')
                        .map(d => d.trim())
                        .filter(d => d);
                }
        }

		for(var i=0;i<badges.length;i++){
			let badge = [];
			badge = badges[i]
					.split('_')
					.map(d => d.trim())
                    .filter(d => d);
					
			if(badge.length == 3){
				if(badge[0] != employeeid)
					tempbadges.push(badges[i]);
			}
		}
		let bs = "";
		
		for(var i=0;i<tempbadges.length;i++){
			if(i> 0)
				bs += ',' + tempbadges[i];
			else 
				bs += tempbadges[i];
		}
		
		this.task.badges = bs;
	}
	
	draw_assigned_badges(){
				
		let r = (this.height - 4)/2
		let y = this.$bar.getY() + this.height/2; 
		let x = this.$bar.getX() + 10 + r;
		
	
		this.bar_group.querySelectorAll('circle.bar-badge').forEach(el => el.remove());
		this.bar_group.querySelectorAll('text.bar-badge-label').forEach(el => el.remove());
		let badges = [];
		// badges: employee
        if (typeof this.task.badges === 'string' || !this.task.badges) {
            if (this.task.badges) {
                    badges = this.task.badges
                        .split(',')
                        .map(d => d.trim())
                        .filter(d => d);
                }
        }
//		this.barbadges  = [];
//		this.barbadgelabeles = [];
		
		for(var i=0;i<badges.length;i++){
			let badge = [];
			badge = badges[i]
					.split('_')
					.map(d => d.trim())
                    .filter(d => d);
					
			if(badge.length == 3){
				let bar_badge = createSVG('circle', {
							cx: x,
							cy: y,
							r: r,
							adid: badge[0],
							class: 'bar-badge',
							append_to: this.bar_group
						});
						
				let bar_badge_label = createSVG('text', {
							x: x,
							y: y,
							adid: badge[0],
							innerHTML: badge[1],
							class: 'bar-badge-label',
							append_to: this.bar_group
						});	
						
//				this.barbadges.push(bar_badge);
//				this.barbadgelabeles.push(bar_badge_label);
				
				x=x+2*r+2;
			
			/*	bar_badge.addEventListener('click', function(e){
					var htmlstr = '<div><input type="button" class="pop_content" taskid="' +this.task.id+ '" badgeid="'  +badge[0]+'" onclick="gantt_removebadge(this);" value="Remove"></input><input type="button" class="pop_content" onclick="hidepopup();" value="Close"></input></div>'
							this.gantt.show_popup({
							target_element: bar_badge,
							title: this.task.name + '(' + badge[2]+ ')',
							subtitle: htmlstr,
							task: this.task,
							e:e
						});	
					
				});  */
			} 
			
		}  
		
	}
	
	draw_issue_icon(){
		this.bar_group.querySelectorAll('text.task_issue_icon').forEach(el => el.remove());
		if(this.task.issuetype && this.task.issuetype !=''){
			let r = (this.height - 4)/2
			let y = this.$bar.getY() + 2*r-4; 
			let x = this.$bar.getX() + 10 + r;
			this.bar_group.querySelectorAll('circle.bar-badge').forEach(function(){
				x += 2*r;
			})
			
			createSVG('text', {
					x: 2 + x,
					y: y,
					fontsize: 24,
					innerHTML: '&#xf06a',
					class: 'task_issue_icon task_issue_icon_' + this.task.issuetype.toLowerCase(),
					append_to: this.bar_group
			});				
		
		}		
	}
	
	draw_progress_bar() {
        if (this.invalid) return;
        this.$bar_progress = createSVG('rect', {
            x: this.x,
            y: (this.y + (this.height * 20)/100),
            width: this.progress_width,
            height: (this.height * 60)/100,
            rx: this.corner_radius,
            ry: this.corner_radius,
            class: 'bar-progress' + !this.task.progressbarclass ? this.task.progressbarclass : '',
            append_to: this.bar_group
        });

        animateSVG(this.$bar_progress, 'width', 0, this.progress_width);
    }

    draw_label() {
        createSVG('text', {
            x: this.x + this.width / 2,
            y: this.y + this.height / 2,
            innerHTML: this.task.name,
            class: 'bar-label',
            append_to: this.bar_group
        });
        // labels get BBox in the next tick
        requestAnimationFrame(() => this.update_label_position());
    }

    draw_actual_label() {
        createSVG('text', {
            x: this.actual_bar_x +5, // this.compute_actual_x() + this.actualwidth / 2,
            y: this.actual_bar_y + ((this.height * 70)/100) / 2,
            innerHTML: this.task.actuallabel? this.task.actuallabel : '',
            class: 'actual-bar-label',
            append_to: this.bar_group
        });
        // labels get BBox in the next tick
       // requestAnimationFrame(() => this. update_actual_bar_label_position());
    }
	
	draw_progress_label() {
		
        createSVG('text', {
            x: this.x + (this.progress_width > 30? this.progress_width -30 : this.progress_width/2),
            y: this.y + this.height / 2,
            innerHTML: this.task.progresslabel? this.task.progresslabel : '',
            class: 'bar-grogress-label',
            append_to: this.bar_group
        });
        // labels get BBox in the next tick
        requestAnimationFrame(() => this.update_label_position());
    }

    draw_resize_handles() {
        if (this.invalid) return;

        const bar = this.$bar;
        const handle_width = 3;
		// draw right 
        createSVG('rect', {
            x: bar.getX() + bar.getWidth()- handle_width-1,
            y: bar.getY() + 1,
            width: handle_width,
            height: this.height - 2,
            rx: this.corner_radius,
            ry: this.corner_radius,
            class: 'handle right',
            append_to: this.handle_group
        });

        createSVG('rect', {
            x: bar.getX() -handle_width + 1,
            y: bar.getY() + 1,
            width: handle_width,
            height: this.height - 2,
            rx: this.corner_radius,
            ry: this.corner_radius,
            class: 'handle left',
            append_to: this.handle_group
        });

        if (this.task.progress && this.task.progress < 100) {
            this.$handle_progress = createSVG('polygon', {
                points: this.get_progress_polygon_points().join(','),
                class: 'handle progress',
                append_to: this.handle_group
            });
        }
    }

    get_progress_polygon_points() {
        const bar_progress = this.$bar_progress;
        return [
            bar_progress.getEndX()-3,
            bar_progress.getY() + bar_progress.getHeight(),
            bar_progress.getEndX() + 3,
            bar_progress.getY() + bar_progress.getHeight(),
            bar_progress.getEndX(),
            bar_progress.getY() + bar_progress.getHeight() - 8.66
        ];
    }

    bind() {
        if (this.invalid) return;
		if(this.gantt.options.custom_popup_html == '') return;
        this.setup_click_event();
    }

    setup_click_event() {
        $.on(this.group, 'focus ' + this.gantt.options.popup_trigger, e => {
            if (this.action_completed) {
                // just finished a move action, wait for a few seconds
                return;
            }

            if (e.type === 'click') {
                this.gantt.trigger_event('click', [this.task]);
            }

            this.gantt.unselect_all();
            this.group.classList.toggle('active');

            this.show_popup(e);
        });
    }

    show_popup(e) {
	//	console.log(e,this.gantt.bar_being_dragged )
		
        if (this.gantt.bar_being_dragged) return;
		
		if(this.gantt.options.custom_popup_html != null && this.gantt.options.custom_popup_html != '') 
		{
			this.gantt.show_popup({
				target_element: this.$bar,
				title: this.task.name + '(' + this.task.resource.name + ')',
				subtitle: this.gantt.showsimplepopup==='yes'? this.task.duration + "(" + subtitle + ")" : this.gantt.options.custom_popup_html,
				task: this.task,
				e:e
			});
			
		}
		else{
			const start_date = date_utils.format(this.task._start, 'MMM D HH mm');
			const end_date = date_utils.format(
				date_utils.add(this.task._end, -1, 'second'),
				'MMM D HH mm'
			);
			const subtitle = start_date + ' - ' + end_date;
			
			let htmlstr = '<div class = "popup_content_section">';
			let parentresource = null;
			
			if(this.task.resource.parentid != "" && this.task.resource.type == 'machine'){
				for(var i=0;i<this.gantt.resources.length;i++){
					if(this.gantt.resources[i].id === this.task.resource.parentid ){
						parentresource = this.gantt.resources[i];
						break;
					}
				}			
			} else 
				parentresource = this.task.resource;
			
			let machines = [];
			if(parentresource != null)
				for(var i=0;i<this.gantt.resources.length;i++){
					if(this.gantt.resources[i].parentid === parentresource.id ){
							machines.push(this.gantt.resources[i]);					
						}
				}			
			
			if(this.task.resource.type == 'line')
				htmlstr += '<div class="pop_content"><span>Production Line:</span><input hidden type="text" class="popup_resource_productionlineno" value="'+parentresource._index+'"><span>'+ parentresource.name +'</span></div>'; 
			else{
				htmlstr += '<div class="pop_content"><span>Work Center:</span><input hidden type="text" class="popup_resource_workcenter" value="'+parentresource._index+'"><span>'+ parentresource.name +'</span></div>'; 
				htmlstr += '<div class="pop_content"><span>Machine:</span><select class="popup_resource_machine" id="pop_resource_machine_selector"  ' +  (!this.task.isendviewonly? '': 'disabled')  +' >';
				
				var subhtmlstr = '';
				let assigned = false;
				for(var i=0;i<machines.length;i++){
					if(this.task.resource.id === machines[i].id ){
						subhtmlstr += '<option value="'+ machines[i]._index +'" selected>'+ machines[i].name+'</option>'	
						assigned = true;
					}
					else
						subhtmlstr += '<option value="'+ machines[i]._index +'" >'+ machines[i].name+'</option>'	
					
				}
				
				if(assigned){
					htmlstr += 	'<option value="" >-</option>' + subhtmlstr;			
				} else
					htmlstr += 	'<option value="" selected >-</option>' + subhtmlstr;		
			
				htmlstr += '<select></div>'
			
			}
			htmlstr += '<div class="pop_content"><span>Start:<span><input type="text" class="popup_task_start" ' +  (!this.task.isstartviewonly? '': 'disabled')  +' value="'+  date_utils.format(this.task._start,'YYYY-MM-DD HH:mm:ss') +'"></div>';
			htmlstr += '<div class="pop_content"><span>End:<span><input type="text" class="popup_task_end" ' +  (!this.task.isendviewonly? '': 'disabled')  +'  value="'+ date_utils.format(this.task._end,'YYYY-MM-DD HH:mm:ss') +'"></div>';
			
			if(this.task.resource.type != 'line'){
				let badges = [];
				// badges: employee
				if (typeof this.task.badges === 'string' || !this.task.badges) {
					if (this.task.badges) {
							badges = this.task.badges
								.split(',')
								.map(d => d.trim())
								.filter(d => d);
						}
				}
				
				if(badges.length > 0){
				//	htmlstr +=  '<hr>'
					htmlstr +=  '<div>Assigned Employees</div>'
				}
				
				for(var i=0;i<badges.length;i++){
					let badge = [];
					badge = badges[i]
							.split('_')
							.map(d => d.trim())
							.filter(d => d);
				
					if(badge.length == 3){
						htmlstr += '<div><input class="task_assigned_employee" type="checkbox" id="E_'+badge[0]+'" name="E_'+badge[0]+ '" checked>';
						htmlstr += '<label for="'+badge[0]+'">'+badge[2]+ '</label></div>';
					
					}
					
				}			
				
			}
			
			
			htmlstr += '<div class="popup_actions"><input type="button" class="pop_content" id="popup_actions_ok" taskid="' +this.task.id+ '" orderno="' + this.task.orderno + '" ordertype="'+this.task.ordertype +'" operation="'+this.task.operation+'" onclick="popupchangedata(this)" value="OK"  ' +  (!this.task.isendviewonly? '': 'disabled')  +' ><input type="button" class="pop_content" onclick="hidepopup();" value="Close"></input>';
			
			if(this.task.resource.type == 'line'){
				htmlstr += '<input type="button" class="pop_content" id="popup_actions_detail" taskid="' +this.task.id+ '" orderno="' + this.task.orderno + '" ordertype="'+this.task.ordertype +'"  onclick="popup_go_to_detail(this)" value="DETAIL" >';  
				//currentclass="' + this.$bar.getAttribute('class') + '"
				
				if(this.$bar.getAttribute("barstatus") != '4' )
					htmlstr += '<input type="button" class="pop_content" id="popup_actions_detail" taskid="' +this.task.id+ '" orderno="' + this.task.orderno + '" ordertype="'+this.task.ordertype +'" onclick="popup_change_order_status(this)"  targetstatus="held" value="HOLD" >';
				
				if(this.$bar.getAttribute("barstatus") == '4' )
					htmlstr += '<input type="button" class="pop_content" id="popup_actions_detail" taskid="' +this.task.id+ '" orderno="' + this.task.orderno + '" ordertype="'+this.task.ordertype +'" onclick="popup_change_order_status(this)" targetstatus="release" value="RELEASE" >';	
			}
			else{


			}
			htmlstr += '</div></div>'
		//	console.log(htmlstr)
			this.gantt.show_popup({
				target_element: this.$bar,
				title: this.task.name + '(' + this.task.resource.name + ')',
				subtitle: this.gantt.showsimplepopup==='yes'? this.task.duration + "(" + subtitle + ")" : htmlstr,
				task: this.task,
				e:e
			});
		}
    }
	
	update_bar_height(){
		//const skipvalidateparent = true
	//	console.log(this.y,this.height)
		const bar = this.$bar;
		this.update_attr(bar, 'y', this.y);
		this.update_attr(bar, 'height', this.height);
		
        this.update_label_position();
        this.update_handle_position();
        this.update_progressbar_position();
		this.update_progress_label_position();
        this.update_arrow_position();
		this.draw_assigned_badges();
		this.draw_issue_icon();
		
	/*	if(!skipvalidateparent)
			this.update_parent_task();
		
		if(!skipsubtask)
			this.update_sub_tasks(mode);	*/	
		
	}

	
    update_bar_position({ x = null, width = null, start = null, end = null, skipvalidateparent = false, skipsubtask = false }) {
        const bar = this.$bar;
	//	console.log(start,end)
		let oldx = bar.getX();
		let oldendx = bar.getEndX();
		
	//	console.log(x,width,start,end,skipvalidateparent,skipsubtask,oldx,oldendx,this.task.isstartviewonly,this.task.isendviewonly);
		
		if(this.task.isstartviewonly && this.task.isendviewonly)
			return;
		

	//	console.log(this.validate_sub_tasks({x: x,width: width,start: start,end: end}))
		if(this.validate_sub_tasks({x: x,width: width,start: start,end: end}) == false)
			return;

		if(start && end){
			let newx = this.get_x_by_date(start);
			let newend = this.get_x_by_date(end);
		//	console.log(newx,newend)
			
			if(newend < newx)
				return;
			
			if(newx != this.x && this.task.isstartviewonly)
				return;
			
			if(newend != (this.x+this.width)  && this.task.isendviewonly)
				return;
			
			const newxs = this.task.dependencies.map(dep => {
					//if(this.gantt.is_task(dep))
						return this.gantt.get_bar(dep).$bar.getX();
					
				});
			
			const newvalid_x = newxs.reduce((prev, curr) => {
					return newx >= curr;
				}, newx);
				if (!newvalid_x) {
					width = null;
					return;
				}
				this.update_attr(bar, 'x', newx);
				this.update_attr(bar, 'width', (newend - newx));
				this.finaldx = 1;
				///this.x = newx;   // do need add this?
				//this.width = (newend - newx);
		}
		else {
			
			if(x && x != this.$bar.getX() && this.task.isstartviewonly)
				return;
			
		//	console.log((x? x: this.$bar.getX())+width, this.$bar.getEndX())
			if(width && this.task.isendviewonly)
			{
				if(this.$bar.getEndX() != ((x? x: this.$bar.getX())+width))
					return;					
			}
			
			var barwidth = this.$bar.getEndX() - this.$bar.getX();
			
			if((x && x == this.$bar.getX() && width && width == barwidth) || (x && x == this.$bar.getX() && !width) || (!x && width && width == barwidth) || (!x && !width))
				return;
			
			if (x) {
				if(skipvalidateparent == false){
					// get all x values of parent task
					const xs = this.task.dependencies.map(dep => {
					//	if(this.gantt.is_task(dep))
							return this.gantt.get_bar(dep).$bar.getX();
					});
					// child task must not go before parent
					const valid_x = xs.reduce((prev, curr) => {
						return x >= curr;
					}, x);
					if (!valid_x) {
						width = null;
						return;
					}
				}
				this.update_attr(bar, 'x', x);
			//	this.x = x;
			}
			if (width && width >= this.gantt.options.column_width) {
				this.update_attr(bar, 'width', width);
				//this.width = width;
			}
			
			
		}
		let latestx = bar.getX();
		let latestendx = bar.getEndX();
		let newwidth = bar.getWidth();
		
		let {newstart,newend} =this.compute_start_end_date();
		this.task._start = newstart;
		this.task._end = newend;
		
		let mode = ""
		
		if(oldx != latestx)
			mode += "S";
		
		if(oldendx != latestendx )
			mode += "E";
		
        this.update_label_position();
        this.update_handle_position();
        this.update_progressbar_position();
		this.update_progress_label_position();
        this.update_arrow_position();
		this.draw_assigned_badges();
		this.draw_issue_icon();
		
		if(!skipvalidateparent)
			this.update_parent_task();
		
		if(!skipsubtask)
			this.update_sub_tasks(mode);
		
		this.gantt.check_overlap_for_resource(this.task.resource.id);
    }
	update_parent_task(){
		if(!this.datechangedbars)
			this.datechangedbars = [];
		
	//	console.log(this.task.parenttask)
		if(!this.task.parenttask || this.gantt.options.updateparent === 'no')
			return;

		if(this.task.parenttask == '')
			return;
		let parenttaskbar = this.gantt.get_bar(this.task.parenttask);
		
		if(!parenttaskbar)
			return;
		
	//	console.log(parenttaskbar)
		
		let endx = this.$bar.getEndX();
		let barx = this.$bar.getX();
		
	//	console.log(parenttaskbar.$bar.getX(),parenttaskbar.$bar.getEndX(), barx,endx );
		
		if(this.task.parenttask.isstartviewonly)
			return;
		
		
		if(parenttaskbar.$bar.getX() > barx || parenttaskbar.$bar.getEndX() < endx){
			let x  = parenttaskbar.$bar.getX() > barx ?  barx : parenttaskbar.$bar.getX();
			let width = parenttaskbar.$bar.getEndX() > endx ? (parenttaskbar.$bar.getEndX() - x) : (endx - x)
			parenttaskbar.update_bar_position({x: x, width:width, skipsubtask: true});
			//this.datechangedbars.push(parenttaskbar);
			parenttaskbar.date_changed()
		}

	}

	validate_sub_tasks({ x = null, width = null, start = null, end = null}){
		let newx = this.$bar.getX();
		let newend = this.$bar.getEndX();
		
		if(start && end){
				newx = this.get_x_by_date(start);
				newend = this.get_x_by_date(end);
		}
		else{
			if(x)
				newx = x;
				
			if(width)	
				newend = newx + width;	
		}
		
		for(var i=0;i<this.task.subtasks.length;i++){

				let id = this.task.subtasks[i]
				let subtask = this.gantt.get_task(id);
				let subtaskbar = this.gantt.get_bar(id);				
				
				if(!subtask == false){
			//		console.log(subtask, subtaskbar.$bar.getX(),subtaskbar.$bar.getEndX(), newx,newend,subtask.isstartviewonly,subtask.isendviewonly)
					if((newx > subtaskbar.$bar.getX()) && subtask.isstartviewonly == true)
						return false;
					
					if(newend < subtaskbar.$bar.getEndX() && subtask.isendviewonly == true)
						return false;
					
					if(subtaskbar.validate_sub_tasks({x: x,width: width,start: start,end: end}) == false)
						return false;
				}				
		}
	//	console.log("end") 
		return true;
	}
	
	update_sub_tasks(mode){
		let subtaskids = [];
		let subtaskbars = [];
		let endx = this.$bar.getEndX();
		let barx = this.$bar.getX();
		
		let gap = this.$bar.getX() -this.x;
		//console.log(gap,mode);
	
		if( mode == "" && gap == 0 )   // when parent move x or move endx to expand the bar, bypass the child
			return;
	
		if(!this.datechangedbars)
			this.datechangedbars = [];
			
		//	console.log(this.task.subtasks, gap)
			this.task.subtasks.map(id =>{
				let subtask = this.gantt.get_task(id);
				let subtaskbar = this.gantt.get_bar(id);
				
				if(!subtask.isstartviewonly){					
					
				//	console.log(subtask);
					if(!subtaskbar == false){
						
						if(mode == "S"){  // change the x
						//	console.log(barx,subtaskbar.$bar.getX() )
							if(barx > subtaskbar.$bar.getX()){
								let subtaskwidth = subtaskbar.$bar.getEndX() -barx;
								subtaskbar.update_bar_position({x: barx, width:subtaskwidth, skipvalidateparent:true});
								subtaskbar.date_changed()
							}							
						}
						else if(mode =="E"){  // change the end of x

							if(endx < subtaskbar.$bar.getEndX()){
								let width = endx - subtaskbar.$bar.getX();
								subtaskbar.update_bar_position({width: width, skipvalidateparent:true});
							}
						}
						else if(mode== "SE"){  // move the subtask  
						
							let x = gap + subtaskbar.x;				
							
							let width = subtaskbar.$bar.getWidth();
					//		console.log(gap,subtaskbar.x, x, subtaskbar.$bar.getX())
							
							if(x != subtaskbar.$bar.getX()){								
								subtaskbar.update_bar_position({x: x, skipvalidateparent:true});
								subtaskbar.date_changed()

							}	
						}
					}
				}
			})				
		
	}
	
	update_bar_resource(diffindex = null){
		const bar = this.$bar;
		let changed = false;
		
		if(this.task.isresourceviewonly)
			return;
		
		if(diffindex){
			const newindex = this.task._originalindex + diffindex;
		//	console.log(newindex, this.task._index, this.task._originalindex)
		

			if(newindex != this.task._index && newindex > -1 && newindex < this.gantt.resources.length && this.task.changeresource )
			{
				let originalresource = this.task.resource
				let newresource = this.gantt.resources[newindex]
				
				//console.log(newresource, originalresource)
				if((newresource.parentid === originalresource.id && newresource.type == "machine" ) ||    // move from work center to machine 
					(newresource.parentid === originalresource.parentid && newresource.type == 'machine' ) ||  //move between machines 
					(newresource.id === originalresource.parentid && originalresource.type == "machine") || // move from machine to workcenter
					this.gantt.options.resourcechange === 'yes'  ){ //|| this.task.resource.taskchange == 'yes'){		// free move
				
					changed = true;
					this.task._index = newindex;
					
					//this.prepare_values();
					this.y = this.gantt.options.header_height +
						this.gantt.options.padding +
						newindex * (this.height + this.gantt.options.padding) 
					//console.log(newindex, this.y)
					this.update_attr(bar, 'y', this.y);
					this.update_label_position();
					this.update_handle_position();
					this.update_progressbar_position();
					this.update_progress_label_position();
					this.update_arrow_position();
					this.draw_assigned_badges();
					this.draw_issue_icon();
					
					this.gantt.check_overlap_for_resource(this.task.resource.id);
					
					this.gantt.trigger_event('resource_change', [
						this.task,
						originalresource,
						newresource
					])					
				}
			}			
		}
		
		
	}
	
    date_changed() {
        let changed = false;
        const { new_start_date, new_end_date } = this.compute_start_end_date();

        if (Number(this.task._start) !== Number(new_start_date)) {
            changed = true;
            this.task._start = new_start_date;
        }
		
        if (Number(this.task._end) !== Number(new_end_date)) {
            changed = true;
            this.task._end = new_end_date;
        }

        if (!changed) return;

		this.task.lastduration = this.task.duration;
		this.task.duration = this.gantt.calculate_task_duration(this.task._start, this.task._end, this.task.resource);
		/*if(this.task.duration < this.task.lastduration){
			const bar = this.$bar;
			this.update_attr(bar, 'fill', 'red');
		} */

		
        this.gantt.trigger_event('date_change', [
            this.task,
            new_start_date,
			new_end_date
            //date_utils.add(new_end_date, -1, 'second')
        ]);
		// date change for the subtasks
		if(this.datechangedbars){
			let subtaskbars = [];
			this.datechangedbars.map(bar => { 
				let count = 0;
				for(var i=0;i<subtaskbars.length;i++){
					if(subtaskbars[i] == bar){
						count =1;
						break;
					}
					
					if(count ==0)
						bar.date_changed();
				}			
			
			})		
			
		}
		this.datechangedbars = [];
    }


	resource_change(){
		let changed = false;
		
		let originalindex = this.task._originalindex;
		let originalresource = this.task.resource
		//if(Number(this.task.subresource) != Number(new_sub_resource)){
		let newresource = this.gantt.resources[this.task._index];
	//	console.log(originalresource,newresource)
		if(this.task._originalindex != this.task._index  && this.task.changeresource &&
			(newresource.parentid === originalresource.id || (newresource.parentid === originalresource.parentid && originalresource.parentid != '' ) || newresource.id === originalresource.parentid || this.gantt.options.resourcechange === 'yes' || this.task.resource.taskchange == 'yes')
			){
			changed = true;
			//this.task.subresource = new_sub_resource;
			this.task._originalindex = this.task._index;
			this.task.resource = newresource;
			
	/*		if(this.task.resource.parentid === originalresource.id && originalresource.parentid == ''){
				let newtask = this.task;
				newtask.id = this.task.id + "_"+ "parent";
				newtask._index = originalindex;
				newtask.index = originalindex;
				newtask.resource = originalresource;
				newtask.resourceid = originalresource.id;
				newtask.sequence = this.gantt.tasks.length;
				newtask.barclass = 'parent-bar';
				newtask.processbarclass = 'parent-progress-bar';
				newtask.changeresource = false;
			//	console.log(newtask)
			
				for(var i=0;i<this.gantt.tasks.length;i++){
					if(this.gantt.tasks[i].id == this.task.id){
						this.gantt.tasks[i]._originalindex = this.task._originalindex ;
						this.gantt.tasks[i]._index = this.task._index;
						this.gantt.tasks[i].index = this.task._index;
						this.gantt.tasks[i].resource = this.task.resource;
						this.gantt.tasks[i].resourceid = this.task.resource.id
						console.log(this.gantt.tasks,this.task, i)
						break;
					}
				}
				this.gantt.tasks.push(newtask);
				this.gantt.render();
			--	const bar = new Bar(this.gantt, newtask);
			--	console.log(bar)
			--	this.gantt.layers.bar.appendChild(bar.group);				
			--	this.gantt.bars.push(bar) 
			
				
			}else if(originalresource.parentid === this.task.resource.id && this.task.resource.parentid == ''){
				let parenttask = this.gantt.get_task(this.task.id + "_"+ "parent");
				var parentbar = document.querySelector(".bar-wrapper[data-id='" + this.task.id + "_parent']");
				console.log(parentbar);
				parentbar.remove();
			} */
		}
		
		
		if(!changed) return;
		
		this.gantt.trigger_event('resource_change', [
			this.task,
			originalresource,
			this.gantt.resources[this.task._index]
		])
		
	}
	
    progress_changed() {
        const new_progress = this.compute_progress();
        this.task.progress = new_progress;
        this.gantt.trigger_event('progress_change', [this.task, new_progress]);
    }

    set_action_completed() {
        this.action_completed = true;
        setTimeout(() => (this.action_completed = false), 1000);
    }

    compute_start_end_date() {
        const bar = this.$bar;
        const x_in_units = 60.00 * bar.getX() / this.gantt.options.column_width;
        const new_start_date = date_utils.add(
            this.gantt.gantt_start,
            x_in_units * this.gantt.options.step,
            'minute'
        );
        const width_in_units = 60.00 * bar.getWidth() / this.gantt.options.column_width;
        const new_end_date = date_utils.add(
            new_start_date,
            width_in_units * this.gantt.options.step,
            'minute'
        );

        return { new_start_date, new_end_date };
    }

    compute_progress() {
        const progress =
            this.$bar_progress.getWidth() / this.$bar.getWidth() * 100;
        return parseInt(progress, 10);
    }

   get_x_by_date(date) {
        const { step, column_width } = this.gantt.options;
        
        const gantt_start = this.gantt.gantt_start;
		
        const diff = date_utils.diff(date, gantt_start, 'minute');
        let x = diff / step * column_width /60.00;

        if (this.gantt.view_is('Month')) {
            const diff = date_utils.diff(task_start, gantt_start, 'day');
            x = diff * column_width / 30;
        }

        return x;
    }

    compute_x() {
        const { step, column_width } = this.gantt.options;
        const task_start = this.task._start;
        const gantt_start = this.gantt.gantt_start;
		
		
        const diff = date_utils.diff(task_start, gantt_start, 'minute');
        let x = diff / step * column_width / 60.00;
		
	//	console.log(gantt_start, task_start, diff, x)
		
        if (this.gantt.view_is('Month')) {
            const diff = date_utils.diff(task_start, gantt_start, 'day');
            x = diff * column_width / 30;
        }
	//	console.log(x)
        return x;
    }

    compute_y() {
	//	console.log(this.task, this.task._index)
        return (
            this.gantt.options.header_height +
            this.gantt.options.padding +
            this.task._index * (this.height + this.gantt.options.padding)
        );
    }
	
    compute_actual_x() {
        const { step, column_width } = this.gantt.options;
        const task_actualstart = this.task._actualstart;
        const gantt_start = this.gantt.gantt_start;
		const gantt_x_offset  = !this.gantt.options.xoffset? this.gantt.options.xoffset : 0;
		

        const diff = date_utils.diff(task_actualstart, gantt_start, 'minute');
        let x = gantt_x_offset + diff / step * column_width / 60.00;

        if (this.gantt.view_is('Month')) {
            const diff = date_utils.diff(task_actualstart, gantt_start, 'day');
            x = gantt_x_offset + diff * column_width / 30;
        }
        return x;
    }

    compute_actual_y() {
        return (
            this.gantt.options.header_height +
            this.gantt.options.padding +
            this.task._index * (this.height + this.gantt.options.padding)
        );
    }	

    get_snap_position(dx) {
        let odx = dx,
            rem,
            position;
		
		return dx;
		
        if (this.gantt.view_is('Week')) {
            rem = dx % (this.gantt.options.column_width / 7);
            position =
                odx -
                rem +
                (rem < this.gantt.options.column_width / 14
                    ? 0
                    : this.gantt.options.column_width / 7);
        } else if (this.gantt.view_is('Month')) {
            rem = dx % (this.gantt.options.column_width / 30);
            position =
                odx -
                rem +
                (rem < this.gantt.options.column_width / 60
                    ? 0
                    : this.gantt.options.column_width / 30);
        } else {
            rem = dx % this.gantt.options.column_width;
            position =
                odx -
                rem +
                (rem < this.gantt.options.column_width / 2
                    ? 0
                    : this.gantt.options.column_width);
        }
        return position;
    }

    update_attr(element, attr, value) {
        value = +value;
        if (!isNaN(value)) {
            element.setAttribute(attr, value);
        }
        return element;
    }

	update_assigned_badges(){
		
		let r = (this.height - 4)/2
	
		let x = this.$bar.getX() + 10 + r;
		let y = this.$bar.getY() + this.$bar.getHeight()/2;
	
		
	//	console.log(this.barbadges)
		this.barbadges = this.barbadges.map((barbadge, i) => {
		//	console.log(barbadge)
			barbadge.setAttribute('x', x);
			barbadge.setAttribute('y', y);
			x =x + 2*r + 2;
		});
		
		this.barbadgelabeles = this.barbadgelabeles.map((barbadgelabele,i) =>{
			barbadgelabele.setAttribute('x', x);
			barbadgelabele.setAttribute('y', y);
			x =x + 2*r + 2;			
			
		})  ;
		
			
	}
    update_progressbar_position() {
		
        this.$bar_progress.setAttribute('x', this.$bar.getX());
		this.$bar_progress.setAttribute('y', this.$bar.getY() + this.$bar.getHeight()*20/100  );
        this.$bar_progress.setAttribute(
            'width',
            this.$bar.getWidth() * (this.task.progress / 100)
        );
    }

    update_label_position() {
        const bar = this.$bar,
            label = this.group.querySelector('.bar-label');

        if (label.getBBox().width > bar.getWidth()) {
            label.classList.add('big');
            label.setAttribute('x', bar.getX() + bar.getWidth() + 5);
        } else {
            label.classList.remove('big');
            label.setAttribute('x', bar.getX() + bar.getWidth() / 2);
        }
		
		label.setAttribute('y', bar.getY() + this.$bar.getHeight() / 2);
    }

    update_progress_label_position() {
        const bar = this.$bar,
            label = this.group.querySelector('.bar-grogress-label');

        //if (label.getBBox().width > bar.getWidth()) {
        label.classList.add('big');
        label.setAttribute('x', bar.getX() + (this.progress_width > 60? this.progress_width -60 : this.progress_width/2));
		
		label.setAttribute('y', bar.getY() + this.$bar.getHeight() / 2);
    }

    update_actual_bar_label_position() {
        const bar = this.$bar,
            label = this.group.querySelector('.actual-bar-label');

        //if (label.getBBox().width > bar.getWidth()) {
        label.classList.add('big');
        label.setAttribute('x', this.$actualbar.getX() + 5);
		
		label.setAttribute('y', this.$actualbar.getY() + this.$actualbar.getHeight() / 2);
    }


    update_handle_position() {
        const bar = this.$bar;
        this.handle_group
            .querySelector('.handle.left')
            .setAttribute('x', bar.getX() + 1)
		//	.setAttribute('y',bar.getY() +1);
        this.handle_group
            .querySelector('.handle.right')
            .setAttribute('x', bar.getEndX() - 2)
		//	.setAttribute('y',bar.getY() + 1);
        this.handle_group
            .querySelector('.handle.left')
        //    .setAttribute('x', bar.getX() + 1)
			.setAttribute('y',bar.getY() +1);
        this.handle_group
            .querySelector('.handle.right')
         //   .setAttribute('x', bar.getEndX() - 9)
			.setAttribute('y',bar.getY() + 1);			

		this.handle_group
            .querySelector('.handle.left')
         //   .setAttribute('x', bar.getEndX() - 9)
			.setAttribute('height',bar.getHeight() + 1);
        
		this.handle_group
            .querySelector('.handle.right')
         //   .setAttribute('x', bar.getEndX() - 9)
			.setAttribute('height',bar.getHeight() + 1);

		
        const handle = this.group.querySelector('.handle.progress');
        handle &&
            handle.setAttribute('points', this.get_progress_polygon_points());
			
		
    }

    update_arrow_position() {
        this.arrows = this.arrows || [];
        for (let arrow of this.arrows) {
            arrow.update();
        }
    }
}

class Arrow {
    constructor(gantt, from_task, to_task) {
        this.gantt = gantt;
        this.from_task = from_task;
        this.to_task = to_task;

        this.calculate_path();
        this.draw();
    }

    calculate_path() {
        let start_x =
            this.from_task.$bar.getX() + this.from_task.$bar.getWidth() / 2;

        const condition = () =>
            this.to_task.$bar.getX() < start_x + this.gantt.options.padding &&
            start_x > this.from_task.$bar.getX() + this.gantt.options.padding;

        while (condition()) {
            start_x -= 10;
        }

        const start_y =
            this.gantt.options.header_height +
            this.gantt.options.bar_height +
            (this.gantt.options.padding + this.gantt.options.bar_height) *
                this.from_task.task._index +
            this.gantt.options.padding;

        const end_x = this.to_task.$bar.getX() - this.gantt.options.padding / 2;
        const end_y =
            this.gantt.options.header_height +
            this.gantt.options.bar_height / 2 +
            (this.gantt.options.padding + this.gantt.options.bar_height) *
                this.to_task.task._index +
            this.gantt.options.padding;

        const from_is_below_to =
            this.from_task.task._index > this.to_task.task._index;
        const curve = this.gantt.options.arrow_curve;
        const clockwise = from_is_below_to ? 1 : 0;
        const curve_y = from_is_below_to ? -curve : curve;
        const offset = from_is_below_to
            ? end_y + this.gantt.options.arrow_curve
            : end_y - this.gantt.options.arrow_curve;

        this.path = `
            M ${start_x} ${start_y}
            V ${offset}
            a ${curve} ${curve} 0 0 ${clockwise} ${curve} ${curve_y}
            L ${end_x} ${end_y}
            m -5 -5
            l 5 5
            l -5 5`;

        if (
            this.to_task.$bar.getX() <
            this.from_task.$bar.getX() + this.gantt.options.padding
        ) {
            const down_1 = this.gantt.options.padding / 2 - curve;
            const down_2 =
                this.to_task.$bar.getY() +
                this.to_task.$bar.getHeight() / 2 -
                curve_y;
            const left = this.to_task.$bar.getX() - this.gantt.options.padding;

            this.path = `
                M ${start_x} ${start_y}
                v ${down_1}
                a ${curve} ${curve} 0 0 1 -${curve} ${curve}
                H ${left}
                a ${curve} ${curve} 0 0 ${clockwise} -${curve} ${curve_y}
                V ${down_2}
                a ${curve} ${curve} 0 0 ${clockwise} ${curve} ${curve_y}
                L ${end_x} ${end_y}
                m -5 -5
                l 5 5
                l -5 5`;
        }
    }

    draw() {
        this.element = createSVG('path', {
            d: this.path,
            'data-from': this.from_task.task.id,
            'data-to': this.to_task.task.id
        });
    }

    update() {
        this.calculate_path();
        this.element.setAttribute('d', this.path);
    }
}

class Popup {
    constructor(parent, custom_html) {
        this.parent = parent;
        this.custom_html = custom_html;
        this.make();
    }

    make() {
        this.parent.innerHTML = `
            <div class="title"></div>
            <div class="subtitle"></div>
            <div class="pointer"></div>
        `;

        this.hide();

        this.title = this.parent.querySelector('.title');
        this.subtitle = this.parent.querySelector('.subtitle');
        this.pointer = this.parent.querySelector('.pointer');
    }

    show(options) {
        if (!options.target_element) {
            throw new Error('target_element is required to show popup');
        }
        if (!options.position) {
            options.position = 'left';
        }
        const target_element = options.target_element;

        if (this.custom_html) {
            let html = this.custom_html(options.task);
            html += '<div class="pointer"></div>';
            this.parent.innerHTML = html;
            this.pointer = this.parent.querySelector('.pointer');
        } else {
            // set data
            this.title.innerHTML = options.title;
            this.subtitle.innerHTML = options.subtitle;
            this.parent.style.width = this.parent.clientWidth + 'px';
        }

        // set position
        let position_meta;
        if (target_element instanceof HTMLElement) {
            position_meta = target_element.getBoundingClientRect();
        } else if (target_element instanceof SVGElement) {
            position_meta = options.target_element.getBBox();
        }

        if (options.position === 'left') {
            this.parent.style.left =
                position_meta.x + (position_meta.width + 10) + 'px';
            this.parent.style.top = position_meta.y + 'px';

            this.pointer.style.transform = 'rotateZ(90deg)';
            this.pointer.style.left = '-7px';
            this.pointer.style.top = '2px';
        }

        // show
        this.parent.style.opacity = 1;
		//$('div.popup-wrapper').show();
    }

    hide() {
        this.parent.style.opacity = 0;
		//$('div.popup-wrapper').show();
    }
}

class Gantt {
    constructor(wrapper, resources, tasks, options) {
        this.setup_wrapper(wrapper);
        this.setup_options(options);
		this.setup_resources(resources);
        this.setup_tasks(tasks);
        // initialize with default view mode
		this.initialize_view_mode();
      //  this.change_view_mode();
        this.bind_events();
		this.change_view_mode();
		
		this.completedrawgantt();
    }

    setup_wrapper(element) {
        let svg_element, wrapper_element;

        // CSS Selector is passed
        if (typeof element === 'string') {
            element = document.querySelector(element);
        }

        // get the SVGElement
        if (element instanceof HTMLElement) {
            wrapper_element = element;
            svg_element = element.querySelector('svg');
        } else if (element instanceof SVGElement) {
            svg_element = element;
        } else {
            throw new TypeError(
                'Frappé Gantt only supports usage of a string CSS selector,' +
                    " HTML DOM element or SVG DOM element for the 'element' parameter"
            );
        }

		element.classList.add("gantt_area")
        // svg element
        if (!svg_element) {
            // create it
            this.$svg = createSVG('svg', {
                append_to: wrapper_element,
                class: 'gantt'
            });
        } else {
            this.$svg = svg_element;
            this.$svg.classList.add('gantt');
        }
		
		// add the resource column in left
        this.$resource_container = document.createElement('div');
		this.$resource_container.classList.add('gantt-resource_container');		
		element.appendChild(this.$resource_container)
	
	
        // wrapper element
        this.$container = document.createElement('div');
        this.$container.classList.add('gantt-container');
		this.$container.classList.add('dragscroll');

        const parent_element = this.$svg.parentElement;
        parent_element.appendChild(this.$container);
        this.$container.appendChild(this.$svg);

        // popup wrapper
        this.popup_wrapper = document.createElement('div');
        this.popup_wrapper.classList.add('popup-wrapper');
        this.$container.appendChild(this.popup_wrapper);
		
		// add the resource column in left		
		this.$resource_column_svg = createSVG('svg', {
                append_to: wrapper_element,
                class: 'gantt_resource_column'
			//	width: this.options.xoffset
			})
				
		this.$resource_container.appendChild(this.$resource_column_svg);
		
		
		// add the right panel

        this.$right_container = document.createElement('div');
		this.$right_container.classList.add('gantt-right-container');		
		element.appendChild(this.$right_container)
		
		this.workperiodcolumns = [];
    }

    setup_options(options) {
        const default_options = {
            header_height: 50,
            column_width: 30,
            step: 24,
            view_modes: [
				'10Minutes',
				'Hour',
                'Quarter Day',
                'Half Day',
                'Day',
                'Week',
                'Month',
                'Year'
            ],
            bar_height: 20,
            bar_corner_radius: 3,
            arrow_curve: 5,
            padding: 20,
            view_mode: 'Day',
            date_format: 'YYYY-MM-DD HH:mm:ss',
            popup_trigger: 'dblclick',
            custom_popup_html: null,
            language: 'en',
			xoffset: 0,
			setindex: 'no',
			showactual: 'no',
			todayasdefault: 'yes',
			viewonly: 'no',
			resourccolumnwidth:200,
			resourcechange: 'no',
			resourcetype: "",
			weekend: '6,0',
			outofworkingtime: '0-4,20-24',
			showsimplepopup: 'yes',
			start: '',
			end: '',
			arrowtype: 'd',
			updateparent: 'no'
        };
        this.options = Object.assign({}, default_options, options);
    }

	setup_resources(resources){
		
		this.resources = resources.map((resource,i) => {
		//	console.log(resource, i)
			resource._index = i;
			if(!this.options.setindex){
				resource._index = i;
			}
			else{
				if(this.options.setindex = "yes")
						resource._index = !resource.index ? i : resource.index;
			}

			return resource;
			
		})
	}
	setup_task(task,i){
		            // convert to Date objects
            task._start = date_utils.parse(new Date(task.start));
            task._end = date_utils.parse(new Date(task.end));
		//	console.log(task.start, new Date(task.start),task._start )
            // make task invalid if duration too large
            if (date_utils.diff(task._end, task._start, 'year') > 10) {
                task.end = null;
            }


            // invalid dates
            if (!task.start && !task.end) {
                const today = date_utils.today();
                task._start = today;
                task._end = date_utils.add(today, 1, 'day');
            }

            if (!task.start && task.end) {
                task._start = date_utils.add(task._end, -1, 'day');
            }

            if (task.start && !task.end) {
                task._end = date_utils.add(task._start, 1, 'day');
            }
		//	console.log(task.start, new Date(task.start),task._start )
            // if hours is not set, assume the last day is full day
            // e.g: 2018-09-09 becomes 2018-09-09 23:59:59
            const task_end_values = date_utils.get_date_values(task._end);
            if (task_end_values.slice(3).every(d => d === 0)) {
                task._end = date_utils.add(task._end, 24, 'hour');
            }

            // invalid flag
            /*if (!task.start || !task.end) {
                task.invalid = true;
            }  */
			
			// add the actual task start date and completion date
			if (!task.actualstart)
					task._actualstart = task._start
			else 
				task._actualstart = date_utils.parse(task.actualstart)

			if (!task.actualend)
					task._actualend = task._start
			else
				task._actualend = date_utils.parse(task.actualend)

            // dependencies
            if (typeof task.dependencies === 'string' || !task.dependencies) {
                let deps = [];
                if (task.dependencies) {
                    deps = task.dependencies
                        .split(',')
                        .map(d => d.trim())
                        .filter(d => d);
                }
                task.dependencies = deps;
            }

            // subtasks
            if (typeof task.subtasks === 'string' || !task.subtasks) {
                let subs = [];
                if (task.subtasks) {
                    subs = task.subtasks
                        .split(',')
                        .map(d => d.trim())
                        .filter(d => d);
                }
                task.subtasks = subs;
            }
			
			if(task.changeresource == false)
				task.changeresource = false;
			else
				task.changeresource = true;
			
            // uids
            if (!task.id) {
                task.id = generate_id(task);
            }

		// cache index
			task.resource = "";
			var checkindex = 0
			if(task.resourceid){
				if(task.resourceid != ""){
					for(var n=0; n< this.resources.length; n++){
						if(this.resources[n].id == task.resourceid){
							task._index = this.resources[n]._index;
							task.resource = this.resources[n];
							checkindex = 1;
							break;
						}
					}
				}
			}
			
			if(checkindex === 0){
				if(task.resourceid){ 
					var _resource = {
							index: this.resources.length,
							_index: this.resources.length,
							id: task.resourceid,
							name: task.resourceid,
							parentid: ''
					}
					var _resources =[]
					_resources.push(_resource);
					this.resources = this.resources.concat(_resources);
					task._index = _resource._index;
					task.resource = _resource;					
				
				} else {
					task._index = i;
					if(!this.options.setindex){
						task._index = i;
					}
					else{
						if(this.options.setindex = "yes")
								task._index = task.index;
					}
				}				
			}
			task._sequence = i;
			
			task._originalindex = task._index;
			//task.duration = 0
			if(!task.duration && !(task.duration > 0) && task.duration != undefined){
				task.duration = task.duration;
			}else{
				task.duration = this.calculate_task_duration(task._start, task._end, task.resource);
			}
			
			if(!task.isstartviewonly)
				task.isstartviewonly = false;
			
			if(!task.isendviewonly)
				task.isendviewonly = false;
			
			if(!task.isresourceviewonly)
				task.isresourceviewonly = false;
			
		//	console.log(task.start, new Date(task.start),task._start )
			return task;
	}
	
    setup_tasks(tasks) {
        // prepare tasks
        this.tasks = tasks.map((task, i) => {
		
		//	console.log(task)
		
            return this.setup_task(task,i);
        });

        this.setup_dependencies();
		this.setup_subtasks();
    }

    setup_dependencies() {
        this.dependency_map = {};
        for (let t of this.tasks) {
            for (let d of t.dependencies) {
				//console.log('dep:', t, d)
				if(!this.is_task(d)){
				
					this.dependency_map[d] = this.dependency_map[d] || [];
					this.dependency_map[d].push(t.id);
				}
            }
        }
    }

    setup_subtasks() {
        this.subtask_map = {};
        for (let t of this.tasks) {
            for (let d of t.subtasks) {
				//console.log('dep:', t, d)
				if(!this.is_task(d)){
				
					this.subtask_map[d] = this.subtask_map[d] || [];
					this.subtask_map[d].push(t.id);
				}
            }
        }
    }

    refresh(tasks) {
        this.setup_tasks(tasks);
        this.change_view_mode();
    }
       
	is_task(id){
		for(var i=0;i<this.tasks.length; i++){
			if(this.tasks[i].id == id)
				return true;
		}
		return false;
	}
	
    initialize_view_mode(mode = this.options.view_mode) {
		this.set_gantt_startend_date();
     //   this.update_view_scale(mode);
        window.setTimeout(this.setup_dates(),0);
        window.setTimeout(this.render(),0);
        // fire viewmode_change event
        this.trigger_event('view_change', [mode]);
    }

    change_view_mode(mode = this.options.view_mode) {
		this.get_scroll_position();
        this.update_view_scale(mode);
        window.setTimeout(this.setup_dates(),0);
        window.setTimeout(this.render(),0);
        // fire viewmode_change event
        this.trigger_event('view_change', [mode]);
    }

    update_view_scale(view_mode) {
        this.options.view_mode = view_mode;

		if (view_mode === '30Minutes') {
            this.options.step = 1/2;
            this.options.column_width = 30;
		}else if (view_mode === '10Minutes') {
            this.options.step = 1/6;
            this.options.column_width = 30;
		}else if (view_mode === 'Hour') {
            this.options.step = 1;
            this.options.column_width = 30;
        } else if (view_mode === 'Day') {
            this.options.step = 24;
            this.options.column_width = 38;
        } else if (view_mode === 'Half Day') {
            this.options.step = 24 / 2;
            this.options.column_width = 38;
        } else if (view_mode === 'Quarter Day') {
            this.options.step = 24 / 4;
            this.options.column_width = 38;
        } else if (view_mode === 'Week') {
            this.options.step = 24 * 7;
            this.options.column_width = 140;
        } else if (view_mode === 'Month') {
            this.options.step = 24 * 30;
            this.options.column_width = 120;
        } else if (view_mode === 'Year') {
            this.options.step = 24 * 365;
            this.options.column_width = 120;
        }
    }

    setup_dates() {
        this.setup_gantt_dates();
        window.setTimeout(this.setup_date_values(),0);
    }
	
	set_gantt_startend_date(){
	    this.gantt_start = this.gantt_end = null;
		
		let globalstart = false;
		let globalend = false; 
		
		/* global start and end are set by option */
		if(!(!this.options.start) && this.options.start != ''  ){
			this.gantt_start = date_utils.parse(this.options.start);// date_utils.start_of(new Date(this.options.start), 'day');
			globalstart = true;
		}
		
		if(!this.options.start){
			
		}
		
		if(!(!this.options.end) && this.options.end != ''  ){
			 this.gantt_end = date_utils.parse(this.options.end); // date_utils.start_of(date_utils.add(new Date(this.options.end),1, 'day'));
			 globalend = true;
		}
	//	console.log('global gant date:', this.gantt_start, this.gantt_end)
		if(	globalstart == false || globalend == false){	
			for (let task of this.tasks) {
				// set global start and end date
				if(globalstart && task._start < this.gantt_start && task._end > this.gantt_start ){
					task._start = this.gantt_start
				}
				
				if(globalend && task._end > this.gantt_start){
					task._end = this.gantt_end
				}

				if(globalstart && task._actualstart < this.gantt_start  && task._actualend > this.gantt_start){
					task._actualstart = this.gantt_start
				}
				
				if(globalend && task._actualend > this.gantt_start){
					task._actualend = this.gantt_end
				}

				if(globalstart && task._end < this.gantt_start){
					task._end = this.gantt_start
				}
				
				if(globalend && task._start > this.gantt_start){
					task._start = this.gantt_end
				}

				if(globalstart && task._actualend < this.gantt_start){
					task._actualend = this.gantt_start
				}
				
				if(globalend && task._actualstart > this.gantt_start){
					task._actualstart = this.gantt_end
				}
				
				if ((!this.gantt_start || task._start < this.gantt_start ) && !globalstart) {
					this.gantt_start = task._start;
				}
				if ((!this.gantt_end || task._end > this.gantt_end) && !globalend) { 
					this.gantt_end = task._end;
				}
				// add the actual start and end date
				if ((!this.gantt_start || task._actualstart < this.gantt_start)&& !globalstart)  {
					this.gantt_start = task._actualstart;
				}
				if ((!this.gantt_end || task._actualend > this.gantt_end) && !globalend)  {
					this.gantt_end = task._actualend;
				}

			}
		}
		
	//	console.log('tasks gant date:', this.gantt_start, this.gantt_end)
		if(!globalstart)
			 this.gantt_start = date_utils.start_of(date_utils.today(), 'day');
        else
			this.gantt_start = date_utils.start_of(this.gantt_start, 'day');
		
		if(!globalend)
			this.gantt_end = date_utils.start_of(date_utils.add(date_utils.today(),2, 'day'), 'day');
		else
			this.gantt_end = date_utils.start_of(this.gantt_end, 'day');
		
		
	//	console.log('validated gant date:', this.gantt_start, this.gantt_end)
	//	console.log(this.options.view_mode, this.gantt_start ,this.options.view_mode, this.gantt_end, date_utils.diff(this.gantt_end,this.gantt_start,  'month'), date_utils.diff( this.gantt_end,this.gantt_start, 'day'))
		
		if(date_utils.diff( this.gantt_end, this.gantt_start,'month') > 60 && !this.view_is(['Year']) ){
			this.update_view_scale("Year");
		} else if(date_utils.diff(this.gantt_end,this.gantt_start, 'month') > 30 && (!this.view_is(['Year']) || !this.view_is(['Month'])) ){
			this.update_view_scale("Month");
		} else if(date_utils.diff(this.gantt_end, this.gantt_start,'month') > 6 && (!this.view_is(['Year']) || !this.view_is(['Month']) || !this.view_is(['Day'])) ){
			this.update_view_scale("Day");
		}
		else if(date_utils.diff(this.gantt_end,this.gantt_start, 'month') > 3 && (!this.view_is(['Year']) || !this.view_is(['Month']) || !this.view_is(['Day'] ) || !this.view_is(['Half Day'] )) ){
			this.update_view_scale("Half Day");
		}else if(date_utils.diff( this.gantt_end, this.gantt_start,'month') > 1 && (!this.view_is(['Year']) || !this.view_is(['Month']) || !this.view_is(['Day'] ) || !this.view_is(['Half Day'] ) || !this.view_is(['Quarter Day'] )) ){
			this.update_view_scale("Quarter Day");
		}else if(date_utils.diff( this.gantt_end,this.gantt_start, 'day') > 7 && (!this.view_is(['Year']) || !this.view_is(['Month']) || !this.view_is(['Day'] ) || !this.view_is(['Half Day'] ) || !this.view_is(['Quarter Day'] ) || !this.view_is(['Hour'] ) )){
			this.update_view_scale("Hour");
		}else if(date_utils.diff(this.gantt_end,this.gantt_start,  'day') > 3 && (!this.view_is(['Year']) || !this.view_is(['Month']) || !this.view_is(['Day'] ) || !this.view_is(['Half Day'] ) || !this.view_is(['Quarter Day'] ) || !this.view_is(['Hour'] ) || !this.view_is(['30Minutes'] )) ){
			this.update_view_scale("30Minutes");
		}else 
			this.update_view_scale("10Minutes");			
		
	}

    setup_gantt_dates() {
     /*   this.gantt_start = this.gantt_end = null;

        for (let task of this.tasks) {
            // set global start and end date
            if (!this.gantt_start || task._start < this.gantt_start) {
                this.gantt_start = task._start;
            }
            if (!this.gantt_end || task._end > this.gantt_end) {
                this.gantt_end = task._end;
            }
			// add the actual start and end date
            if (!this.gantt_start || task._actualstart < this.gantt_start) {
                this.gantt_start = task._actualstart;
            }
            if (!this.gantt_end || task._actualend > this.gantt_end) {
                this.gantt_end = task._actualend;
            }

        }

        this.gantt_start = date_utils.start_of(this.gantt_start, 'day');
        this.gantt_end = date_utils.start_of(this.gantt_end, 'day');
		*/
        // add date padding on both sides
		/*
        if (this.view_is(['10Minutes']) || this.view_is(['30Minutes'])) {
            this.gantt_start = date_utils.add(this.gantt_start, -2, 'hour');
            this.gantt_end = date_utils.add(this.gantt_end, 2, 'hour');
        } else if (this.view_is(['Hour'])) {
            this.gantt_start = date_utils.add(this.gantt_start, -3, 'day');
            this.gantt_end = date_utils.add(this.gantt_end, 3, 'day');
        } else	if (this.view_is(['Quarter Day', 'Half Day'])) {
            this.gantt_start = date_utils.add(this.gantt_start, -7, 'day');
            this.gantt_end = date_utils.add(this.gantt_end, 7, 'day');
        } else if (this.view_is('Month')) {
            this.gantt_start = date_utils.start_of(this.gantt_start, 'year');
            this.gantt_end = date_utils.add(this.gantt_end, 1, 'year');
        } else if (this.view_is('Year')) {
            this.gantt_start = date_utils.add(this.gantt_start, -2, 'year');
            this.gantt_end = date_utils.add(this.gantt_end, 2, 'year');
        } else {
            this.gantt_start = date_utils.add(this.gantt_start, -1, 'month');
            this.gantt_end = date_utils.add(this.gantt_end, 1, 'month');
        } */
		
		if(this.options.todayasdefault == 'yes'){
			if( this.gantt_start > date_utils.today())
				this.gantt_start = date_utils.start_of(date_utils.today(),'day')
			
			if(this.gantt_end < date_utils.today())
				this.gantt_end = date_utils.start_of(date_utils.add(date_utils.today(), 1,'day'),'day')
		}
		
	//	console.log('todatasdefault gant date:', this.gantt_start, this.gantt_end)
		
		
		
		if(!this.options.start || !this.options.end){
			if (this.view_is(['10Minutes']) || this.view_is(['30Minutes'])) {
				this.gantt_start = date_utils.add(this.gantt_start, -1, 'hour');
				this.gantt_end = date_utils.diff( this.gantt_end,this.gantt_start, 'month') > 4? date_utils.add(this.gantt_start, 4, 'month'): date_utils.add(this.gantt_end, 1, 'hour');
				
			} else if (this.view_is(['Hour'])) {
				this.gantt_start = date_utils.add(this.gantt_start, -1, 'day');
			//    this.gantt_end = date_utils.add(this.gantt_end, 3, 'day');
				this.gantt_end = date_utils.diff(this.gantt_end,this.gantt_start,  'month') > 12? date_utils.add(this.gantt_start, 12, 'month'): date_utils.add(this.gantt_end, 1, 'day');
			} else	if (this.view_is(['Quarter Day', 'Half Day'])) {
				this.gantt_start = date_utils.add(this.gantt_start, -1, 'day');
			 //   this.gantt_end = date_utils.add(this.gantt_end, 7, 'day');
				this.gantt_end = date_utils.diff( this.gantt_end,this.gantt_start, 'month') > 6? date_utils.add(this.gantt_start, 6, 'month'): date_utils.add(this.gantt_end, 1, 'day');
			} else if (this.view_is('Month')) {
				this.gantt_start = date_utils.start_of(this.gantt_start, 'year');
			//    this.gantt_end = date_utils.add(this.gantt_end, 1, 'year');
				this.gantt_end = date_utils.diff(this.gantt_end, this.gantt_start, 'year') > 10? date_utils.add(this.gantt_start, 10, 'year'): date_utils.add(this.gantt_end, 1, 'month');
			} else if (this.view_is('Year')) {
				this.gantt_start = date_utils.add(this.gantt_start, -2, 'year');
				this.gantt_end = date_utils.add(this.gantt_end, 2, 'year');
			} else {
				this.gantt_start = date_utils.add(this.gantt_start, -1, 'month');
			//    this.gantt_end = date_utils.add(this.gantt_end, 1, 'month');
				this.gantt_end = date_utils.diff(this.gantt_end,this.gantt_start,  'year') > 6? date_utils.add(this.gantt_start, 6, 'year'): date_utils.add(this.gantt_end, 1, 'month');
			}
		}
		
		if(date_utils.diff( this.gantt_end,this.gantt_start, 'day') < 1)
			this.gantt_end = date_utils.add(this.gantt_start, 1, 'day')
	//	console.log('view mode gant date:', this.gantt_start, this.gantt_end)
    }

    setup_date_values() {
        this.dates = [];
        let cur_date = null;

        while (cur_date === null || cur_date < this.gantt_end) {
            if (!cur_date) {
                cur_date = date_utils.clone(this.gantt_start);
            } else {
                if (this.view_is('Year')) {
                    cur_date = date_utils.add(cur_date, 1, 'year');
                } else if (this.view_is('Month')) {
                    cur_date = date_utils.add(cur_date, 1, 'month');
                } else if (this.view_is('10Minutes')) {
                    cur_date = date_utils.add(cur_date, 10, 'minute');
                } else if (this.view_is('30Minutes')) {
                    cur_date = date_utils.add(cur_date, 30, 'minute');
                } else {
                    cur_date = date_utils.add(
                        cur_date,
                        this.options.step,
                        'hour'
                    );
                }
            }
            this.dates.push(cur_date);
        }
    }

    bind_events() {
        this.bind_grid_click();
        this.bind_bar_events();
    }

    render() {
		
        this.clear();
        this.setup_layers();
        window.setTimeout(this.make_grid(),0);
        window.setTimeout(this.make_dates(),0);
        window.setTimeout(this.make_bars(),0);
        window.setTimeout(this.make_arrows(),0);
        window.setTimeout(this.map_arrows_on_bars(),0);
        window.setTimeout(this.set_width(),0);
        window.setTimeout(this.set_scroll_position(),0);
		
		window.setTimeout(this.check_overlap_bar(),0);
		
		
	//	if(!this.workperiodcolumns == false){
	//	console.log(this.workperiodcolumns);
		
		/*if($("g.workperiod-grid-column").length > 0){
			$("g.workperiod-grid-column").each(function(){
				$(this).remove();
			});
		}  */

			let that = this;
			this.workperiodcolumns.forEach(function(item,index){
				that.draw_grid_workperiodcolumn(item.start,item.end,item.type,item.fillcolor,item.code, false);
			})
			
	//	}
	
		 if(dragscroll != 'undefined')
			dragscroll.reset();
		
    }

    setup_layers() {
        this.layers = {};
        const layers = ['grid',  'workperiod','date', 'arrow','highlight', 'progress', 'bar', 'details'];
        // make group layers
        for (let layer of layers) {
            this.layers[layer] = createSVG('g', {
                class: layer,
                append_to: this.$svg
            });
        }
    }

    make_grid() {
        this.make_grid_background();
        window.setTimeout(this.make_grid_rows(),0);
        window.setTimeout(this.make_grid_header(),0);
        window.setTimeout(this.make_grid_ticks(),0);
        window.setTimeout(this.make_grid_highlights(),0);
		window.setTimeout(this.make_grid_weekends(),0);
		window.setTimeout(this.make_grid_outofworktime(),0);
    }

    make_grid_background() {
        const grid_width = this.dates.length * this.options.column_width;
        const grid_height =
            this.options.header_height +
            this.options.padding +
            (this.options.bar_height + this.options.padding) *
				this.resources.length
            //    this.tasks.length;

        createSVG('rect', {
            x: this.options.xoffset,
            y: 0,
            width: grid_width,
            height: grid_height,
            class: 'grid-background',
            append_to: this.layers.grid
        });

        $.attr(this.$svg, {
            height: grid_height + this.options.padding + 100,
            width: '100%'
        });

		
        createSVG('rect', {
            x: 0,
            y: 0,
            width: this.options.resourccolumnwidth,
            height: grid_height,
			fill:'none',
            class: 'resource-grid-background',
            append_to: this.$resource_column_svg
        });
		
		$.attr(this.$resource_column_svg, {
            height: grid_height + this.options.padding + 100,
            width: this.options.resourccolumnwidth
        });
    }

    make_grid_rows() {
        const rows_layer = createSVG('g', { append_to: this.layers.grid });
        const lines_layer = createSVG('g', { append_to: this.layers.grid });
		
		const resource_layer = createSVG('g', { class: 'resource-grid-column', width:this.options.resourccolumnwidth, append_to: this.$resource_column_svg });

        const row_width = this.dates.length * this.options.column_width;
        const row_height = this.options.bar_height + this.options.padding;

        let row_y = this.options.header_height + this.options.padding / 2;
		
		var indexlist = [];
		
	
      /*  for (let resource of this.resources) {
		//	console.log(resource)
			

				
				row_y += this.options.bar_height + this.options.padding;
			}
        }  */
		let options = this.options;
	
		this.resources.forEach(function(resource,index){	
			row_y = options.header_height + options.padding / 2 + (options.bar_height + options.padding) * index;
			const grid_row_layer = createSVG('g', { class: 'resource-grid-row', append_to: rows_layer });
			const resource_row_layer = createSVG('g', { class: 'resource-grid-column', width: options.resourccolumnwidth, append_to: resource_layer });
			
			var addedtask = 0
			if(addedtask == 0){
				asycreateSVG('rect', {
					x: options.xoffset,
					y: row_y,
					width: row_width,
					height: row_height,
					class: 'grid-row',
					append_to: grid_row_layer
				});

				asycreateSVG('line', {
					x1: options.xoffset,
					y1: row_y + row_height,
					x2: row_width,
					y2: row_y + row_height,
					class: 'row-line',
					append_to: grid_row_layer //lines_layer
				});


				// create the resource
			/*	asycreateSVG('rect', {
					x: 0,
					y: row_y,
					width: 15,
					height: row_height,
					class: (( resource.type =='line')? 'resource-row-line'  :(( resource.parentid =='' || resource.type =='workcenter') ? 'resource-row-workcenter': 'resource-row-equipment')),
					parentnode: resource.parentid,
					append_to: resource_row_layer //resource_layer
				});
			*/	
				var resourceicon = asycreateSVG('rect', {
					x: 0,
					y: row_y,
					width: row_width-15,
					height: row_height,
					resourcetype: resource.type,
					resource: resource.name,
					class: 'resource-grid-row',
					parentnode: resource.parentid,
					append_to: resource_row_layer //resource_layer
				});
				var offsetx = (resource.type =='line') ?  0 : (resource.type =='workcenter'? 8 : (resource.type =='machine'?  16:4));
				createSVG('text', {
					x: 0 + offsetx,
					y: row_y + row_height/2+5,
					fontsize: 24,
					innerHTML: ( resource.type =='line') ? '&#xf275' : ( resource.type =='workcenter'? '&#xf1b3' : (resource.type =='machine'?  '&#xf085' : '&#xf03e' )),
					class: 'resource-grid-row-text-icon ' + (( resource.type =='line')? 'resource-row-line'  :(resource.type =='workcenter'? 'resource-row-workcenter': 'resource-row-equipment')),
					parentnode: resource.parentid,
					append_to: resource_row_layer //resource_layer
				});		
				
				offsetx = (resource.type =='line') ?  0 : (resource.parentid =='' || resource.type =='workcenter'? 5 :  10);
				
				let resourcename  = resource.name.indexOf('resource_') == 0? resource.name.substring(9, resource.name.length): resource.name;
				
				createSVG('text', {
					x: 30 + offsetx,
					y: row_y + row_height/2,
				//	width: this.options.xoffset,
				//	height: row_height,
					innerHTML: ((resource.type =='line') ? '-' : (resource.parentid =='' || resource.type =='workcenter'? '&nbsp;&nbsp;&nbsp;-' :  '&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;-' )) + resourcename,
					//((resource.parentid =='' || resource.type =='line' || resource.type =='workcenter') ? '': '&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;') +resource.name ,
					class: 'resource-grid-row-text ',
					parentnode: resource.parentid,
					append_to: resource_row_layer //resource_layer
				});
				//(resource.parentid ==''? '- ': '&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;') +
				
				asycreateSVG('line', {
					x1: 0,
					y1: row_y + row_height,
					x2: options.resourccolumnwidth,
					y2: row_y + row_height,
					class: 'row-line',
					append_to: resource_row_layer //resource_layer
				});		
			}
			
		})
    }


    make_grid_header() {
        const header_width = this.dates.length * this.options.column_width;
        const header_height = this.options.header_height + 10;
        asycreateSVG('rect', {
            x: this.options.xoffset,
            y: 0,
            width: header_width,
            height: header_height,
            class: 'grid-header',
            append_to: this.layers.grid
        });
		
	    asycreateSVG('rect', {
            x: 0,
            y: 0,
            width: this.options.resourccolumnwidth,
            height: header_height,
			fill:'none',
            class: 'grid-header',
            append_to: this.$resource_column_svg
        });	
		
    }

    make_grid_ticks() {
        let tick_x = this.options.xoffset;
        let tick_y = this.options.header_height + this.options.padding / 2;
        let tick_height =
            (this.options.bar_height + this.options.padding) *
            this.tasks.length;
		 const ticks_layer = createSVG('g', { append_to: this.layers.grid });
		 
		//console.log(this.dates)
    //    for (let date of this.dates) {
		let options = this.options;
		let view_mode = this.options.view_mode;
		this.dates.forEach(function(date,index){
            let tick_class = 'tick';
            // thick tick for monday
            if (view_mode == 'Day' && date.getDate() === 1) {
                tick_class += ' thick';
            }
            // thick tick for first week
            if (
                (view_mode =='Week') &&
                date.getDate() >= 1 &&
                date.getDate() < 8
            ) {
                tick_class += ' thick';
            }
            // thick ticks for quarters
            if ((view_mode =='Month') && (date.getMonth() + 1) % 3 === 0) {
                tick_class += ' thick';
            }
			
			if(((view_mode =='Quarter Day') || (view_mode =='Half Day') || (view_mode =='Hour')) && date.getHours() == 0){
                tick_class += ' thick';
            }
			
			if((view_mode =='10Minutes') && date.getMinutes() == 0){
                tick_class += ' thick';
            }

            createSVG('path', {
                d: `M ${tick_x} ${tick_y} v ${tick_height}`,
                class: tick_class,
                append_to: ticks_layer //this.layers.grid
            });

            if ((view_mode =='Month')) {
                tick_x +=
                    date_utils.get_days_in_month(date) *
                    options.column_width /
                    30;
            } else {
                tick_x += options.column_width;
            }
        });
    }

    make_grid_highlights() {
        // highlight today's date
        if (this.view_is('Day')|| this.view_is('Quarter Day') || this.view_is('Half Day') || this.view_is('Hour') || this.view_is('30Minutes')  || this.view_is('10Minutes') ) {
            const x = this.options.xoffset + 
                date_utils.diff(date_utils.today(), this.gantt_start, 'hour') /
                this.options.step *
                this.options.column_width;
            const y = 0;

            const width = this.options.column_width *24/this.options.step;
            const height =
                (this.options.bar_height + this.options.padding) *
                    this.resources.length +
                this.options.header_height +
                this.options.padding / 2;

            asycreateSVG('rect', {
                x,
                y,
                width,
                height,
                class: 'today-highlight',
                append_to: this.layers.grid
            });
        }
		
		this.make_line_highlights();
    }

	make_line_highlights(){
		
		
	    const x = this.options.xoffset + 
                date_utils.diff(date_utils.parse(new Date()), this.gantt_start, 'minute') / 60 /
                this.options.step *
                this.options.column_width;
        const y = 0;	

		const height =
                (this.options.bar_height + this.options.padding) *
                    this.resources.length +
                this.options.header_height +
                this.options.padding / 2;
		
		this.nowhighlight = asycreateSVG('line', {
						x1: x,
						y1: 0,
						x2: x,
						y2: height,
						class: 'now-highlight-line',
						innerHTML: '<title class="now-highlight-title">'+ date_utils.parse(new Date()) + '</title>',
						append_to: this.layers.highlight
		});
		/*
		var self = this;

		self.start = function(){
			self.interval = setInterval(function() { self.update_now_highlight(); },1000 * 60 * 1);
		}; */
		if(this.nowhighlightinterval)
			window.clearInterval(this.nowhighlightinterval);
		
		var self = this;
		this.nowhighlightinterval = window.setInterval(function() { self.update_now_highlight(); },1000 * 60 * 1);
		// setInterval(function() { self.update_now_highlight(); },1000 * 60 * 1);
	}
	
	update_now_highlight(){
		
	    const x = this.options.xoffset + 
                date_utils.diff(date_utils.parse(new Date()), this.gantt_start, 'minute') / 60 /
                this.options.step *
                this.options.column_width;
				
			
		var el  = document.querySelector("line.now-highlight-line");
		
	//	console.log(date_utils.parse(new Date()), x, el)	
		if(el == null) return;
		
		el.setAttribute('x1', x);
		el.setAttribute('x2', x);
		el.innerHTML = '<title class="now-highlight-title">'+ date_utils.parse(new Date()) + '</title>';
			
	//	$(el).html(date_utils.parse(new Date()));
	//	window.setTimeout(this.update_now_highlight(), 1*60*1000);
	}

	make_grid_weekends(){
		if (this.view_is('Day') || this.view_is('Quarter Day') || this.view_is('Half Day') || this.view_is('Hour') || this.view_is('30Minutes')  || this.view_is('10Minutes') ) {
			const offset = this.view_is('Hour')? 0:0
			const width = this.options.column_width * 24/this.options.step;
            const height =
                (this.options.bar_height + this.options.padding) *
                    this.resources.length +
                this.options.header_height +
                this.options.padding / 2;
			
			const weekends_layer = createSVG('g', {class: 'weekend-grid-column', append_to: this.layers.grid });
			
			var weekends = [];
			
			if(this.options.weekend != ''){
				weekends =this.options.weekend.split(',') 
			}
		//	console.log(weekends,this.options.weekend);
			if(weekends.length == 0)
				return;
			
			var calculateddate =new Date(new Date().setHours(0,0,0,0));
			let options = this.options;
			let gantt_start = this.gantt_start;
			this.get_dates_to_draw().forEach(function(date,index){
			//for (let date of this.get_dates_to_draw()) {
				//var dayofdate = new Date(date.date);
				//console.log(date, dayofdate, dayofdate.getDay())
				var dayofdate = new Date(new Date(date.date).setHours(0,0,0,0));
				
				if(calculateddate.getDate() != dayofdate.getDate()){
					calculateddate = dayofdate
					var day = dayofdate.getDay();
					var isweekend =false
					for(var i=0;i<weekends.length;i++){
						if(day == weekends[i]){
							isweekend = true;
						}
					}
										
					if(isweekend){
						const x = options.xoffset + 
						(date_utils.diff(dayofdate, gantt_start, 'hour') -offset) /
						options.step *
						options.column_width;
						
						const y = 0;	
						//console.log(dayofdate,x,y,width,height);
						
						asycreateSVG('rect', {
							x,
							y,
							width,
							height,
						//	fill: '#F5B7B1',
							class: 'weekend-highlight',
							append_to: weekends_layer //this.layers.grid					
						});  
					} 
				}				
					
			}); 
		}		
	}
	
	draw_grid_workperiodcolumn(starttime,endtime, type,fillcolor,code = '', newitem = true){
	//	console.log(start,end, type, fillcolor)

		
		if (this.view_is('Day') || this.view_is('Quarter Day') || this.view_is('Half Day') || this.view_is('Hour') || this.view_is('30Minutes')  || this.view_is('10Minutes') ) {

			let start = date_utils.parse(new Date(starttime));
			let end = date_utils.parse(new Date(endtime));
			
			if(start < this.gantt_start || end > this.gantt_end || start > this.gantt_end || end < this.gantt_start)
				return;
			
			//console.log(start,end, type, fillcolor)
			/*
			if(!this.workperiodcolumns){
				let this.workperiodcolumns = [];
			}*/
			
			let options = this.options;
			let gantt_start = this.gantt_start;
			const offset = 0 ; // this.view_is('Hour')? 0:0
            const height =
                (this.options.bar_height + this.options.padding) *
                    this.resources.length +
                this.options.header_height +
                this.options.padding / 2;	
				
			var timelayer = this.layers.workperiod; // createSVG('g', {class: 'worktime-grid-column', append_to: this.layers.grid });
			/*
			if(type == "weekend" && $("g.weekend-grid-column").length >0 )
				 timelayer =  $("g.weekend-grid-column");
			else if(type == "weekend" && $("g.weekend-grid-column").length ==0 )
				 timelayer =  createSVG('g', {class: 'weekend-grid-column', append_to: this.layers.workperiod });
			else if(type == "workoff" &&  $("g.outofworktime-grid-column").length >0)
				 timelayer =  $("g.outofworktime-grid-column");
			else if(type == "workoff")
				 timelayer =  createSVG('g', {class: 'outofworktime-grid-column', append_to: this.layers.workperiod });
			else if($("g.workperiod-grid-column").length >0 )
				timelayer =  $("g.workperiod-grid-column");
			else 
				timelayer =  createSVG('g', {class: 'workperiod-grid-column', append_to: this.layers.workperiod });
			*/
			var x = options.xoffset + 
					(date_utils.diff(start, gantt_start, 'hour') -offset )/
						options.step * 
						options.column_width;
			
			var width = options.column_width * (date_utils.diff(end, start, 'hour'))/options.step;
			
		//	console.log(start,end, type, fillcolor,timelayer)
			
			let y = 0;
			let classname =  'workperiod_grid_column' //(type == 'weekend'? 'weekend-highlight' : (type == 'workoff'? 'outofwork-highlight' :'') );
			let wpcolumn = createSVG('rect', {
									x,
									y,
									width,
									height,
									code: code,
								//	fill: '#FEF9E7',
									class: classname,
									append_to:  timelayer //this.layers.grid					
								});

			if(fillcolor !=''){
				wpcolumn.setAttribute('fill', fillcolor);
				wpcolumn.setAttribute('opacity', "0.1");
			}
			
			if(newitem){
				let newwpcolumn = {
					start: start,
					end: end,
					type: type,
					code, code,
					fillcolor: fillcolor
				};
				
				
				this.workperiodcolumns.push(newwpcolumn);
			}
		}		
	}
	
	make_grid_outofworktime(){
		
		if (this.view_is('Day') || this.view_is('Quarter Day') || this.view_is('Half Day') || this.view_is('Hour') || this.view_is('30Minutes')  || this.view_is('10Minutes') ) {
			const offset = this.view_is('Hour')? 0:0
            const height =
                (this.options.bar_height + this.options.padding) *
                    this.resources.length +
                this.options.header_height +
                this.options.padding / 2;
			
			var outofworkingtimes = []
			if(this.options.outofworkingtime != '')
				outofworkingtimes = this.options.outofworkingtime.split(',')
			
			//console.log(outofworkingtimes,this.options.outofworkingtime)
			
			if(outofworkingtimes.length == 0 )
				return;
			
			var weekends = [];
			
			if(this.options.weekend != ''){
				weekends =this.options.weekend.split(',') 
			}		

			let grid = this.layers.grid;
			const outofworktime_layer = createSVG('g', { class: 'outofworktime-grid-column',append_to: grid });
			
			//console.log(outofworkingtimes, weekends)
			var calculateddate =new Date(new Date().setHours(0,0,0,0));
		//	console.log(this.gantt_start)
			let options = this.options;
			let gantt_start = this.gantt_start;
			
			
			this.get_dates_to_draw().forEach(function(date,index){		
			//for (let date of this.get_dates_to_draw()) {
				// set the date to midnight
				var dayofdate = new Date(new Date(date.date).setHours(0,0,0,0));
			//	console.log(dayofdate,this.gantt_start,date_utils.diff(dayofdate, this.gantt_start, 'hour'))
				
				
				if(calculateddate.getDate() != dayofdate.getDate()){
					calculateddate = dayofdate
					var day = dayofdate.getDay();					
					
					var isweekend =false
					for(var i=0;i<weekends.length;i++){
							if(day == weekends[i]){
								isweekend = true;
							}
					}				
					
					if(!isweekend){
						
						for(var i=0;i<outofworkingtimes.length;i++){
							var outoftimes =outofworkingtimes[i].split("-") 
					//		console.log(outoftimes)
							if(outoftimes.length ==2){
						
								var x = options.xoffset + 
								(date_utils.diff(dayofdate, gantt_start, 'hour') -offset + parseFloat(outoftimes[0]) )/
								options.step * 
								options.column_width;

								var width = options.column_width * (parseFloat(outoftimes[1])-parseFloat(outoftimes[0]))/options.step;
								//var width = this.options.column_width * 24/this.options.step;
								
								var y = 0;	
					//			console.log(dayofdate,outoftimes,x,y,width,height);								
								
								createSVG('rect', {
									x,
									y,
									width,
									height,
								//	fill: '#FEF9E7',
									class: 'outofwork-highlight',
									append_to: outofworktime_layer //this.layers.grid					
								}); 
							}						
						}
					}
				}				
					
			} );
		}
	}

	calculate_task_duration(start, end, resource){
        let taskdates = [];
        let cur_date = null;
		let totalduration = 0;

		var outofworkingtimes = []
		if(this.options.outofworkingtime != '')
				outofworkingtimes = this.options.outofworkingtime.split(',')
		
        while (cur_date === null || cur_date < end) {
            if (!cur_date) {
                cur_date = date_utils.clone(new Date(new Date(start).setHours(0,0,0,0)));
				taskdates.push( date_utils.clone(start));
            } else {
                    cur_date = date_utils.add(
                        cur_date,
                        1,
                        'day'
                    );
					taskdates.push(cur_date > end? end: cur_date);
                }				
                        
        }
	//	console.log(taskdates)
		// check weekend and out of work time
		//let lastdate = start
		if(start.getDate() == end.getDate()){
			let outofminutes = 0
			for(var i=0;i<outofworkingtimes.length;i++){
				var outoftimes =outofworkingtimes[i].split("-")				
				if(outoftimes.length == 2){
					if(start.getHours() <= outoftimes[0] && end.getHours() >= outoftimes[1] ){
						outofminutes += (outoftimes[1] -  outoftimes[0]) * 60;
					}else if (start.getHours() > outoftimes[0] && start.getHours() < outoftimes[1] && end.getHours() > outoftimes[1]){
						outofminutes += (outoftimes[1] -  start.getHours()) * 60;
					}else if (start.getHours() > outoftimes[0] && start.getHours() < outoftimes[1] && end.getHours() < outoftimes[1]){
						outofminutes += date_utils.diff(end,start, 'minutes');
					}					
				}				
			}
			totalduration = date_utils.diff(end,start,'minutes') -	outofminutes			
		}
		else{
			let lastdate = start
			for (let date of taskdates){
			//	console.log(date,totalduration)
				if(date.getDate() === 0 || date.getDate() === 6){
									
				}else{
					let outofminutes = 0
					for(var i=0;i<outofworkingtimes.length;i++){
						var outoftimes =outofworkingtimes[i].split("-")				
						if(outoftimes.length == 2){
							if(date.getHours() <= outoftimes[0]){
								outofminutes += (outoftimes[1] -  outoftimes[0]) * 60;
							}else if (date.getHours() > outoftimes[0] && date.getHours() < outoftimes[1]){
								outofminutes += (outoftimes[1] -  date.getHours()) * 60;
							}						
						}				
					}
			//		console.log(date,totalduration,date_utils.diff(date,lastdate,'minutes'),outofminutes )
					totalduration += date_utils.diff(date,lastdate,'minutes') -	outofminutes
					
				}
				lastdate = date;			
			}
		}
		return totalduration > 0? totalduration:0 ;
		
	}

    make_dates() {
		let datelayer = this.layers.date;
		let grid = this.layers.grid;
		
		this.get_dates_to_draw().forEach(function(date,index){
        //for (let date of this.get_dates_to_draw()) {
            createSVG('text', {
                x: date.lower_x,
                y: date.lower_y,
                innerHTML: date.lower_text,
                class: 'lower-text',
                append_to: datelayer
            });

            if (date.upper_text) {
                const $upper_text = createSVG('text', {
                    x: date.upper_x,
                    y: date.upper_y,
                    innerHTML: date.upper_text,
                    class: 'upper-text',
                    append_to: datelayer
                });

                // remove out-of-bound dates
                if (
                    $upper_text.getBBox().x2 > grid.getBBox().width
                ) {
                    $upper_text.remove();
                }
            }
        });
    }

    get_dates_to_draw() {
        let last_date = null;
        const dates = this.dates.map((date, i) => {
            const d = this.get_date_info(date, last_date, i);
            last_date = date;
            return d;
        });
	//	console.log(dates)
        return dates;
    }

    get_date_info(date, last_date, i) {
        if (!last_date) {
            last_date = date_utils.add(date, 1, 'year');
        }
        const date_text = {
            '10Minutes_lower': date_utils.format(
                date,
                'mm',
                this.options.language
            ),
            '30Minutes_lower': date_utils.format(
                date,
                'mm',
                this.options.language
            ),			
			'Hour_lower': date_utils.format(
                date,
                'HH',
                this.options.language
            ),
            'Quarter Day_lower': date_utils.format(
                date,
                'HH',
                this.options.language
            ),
            'Half Day_lower': date_utils.format(
                date,
                'HH',
                this.options.language
            ),
            Day_lower:
                date.getDate() !== last_date.getDate()
                    ? date_utils.format(date, 'D', this.options.language)
                    : '',
            Week_lower:
                date.getMonth() !== last_date.getMonth()
                    ? date_utils.format(date, 'D MMM', this.options.language)
                    : date_utils.format(date, 'D', this.options.language),
            Month_lower: date_utils.format(date, 'MMMM', this.options.language),
            Year_lower: date_utils.format(date, 'YYYY', this.options.language),
			
            '10Minutes_upper':
                date.getHours() !== last_date.getHours()
                    ? date_utils.format(date, 'D/MM/YYYY HH', this.options.language)
                    : '',
			'30Minutes_upper':
                ((date.getHours() === 0 || date.getHours() === 6 || date.getHours() === 12 || date.getHours() === 18) && date.getHours() !== last_date.getHours())
                    ? date_utils.format(date, 'D/MM/YYYY HH', this.options.language)
                    : '',	
            'Hour_upper':
                date.getDate() !== last_date.getDate()
                    ? date_utils.format(date, 'D MMM YYYY', this.options.language)
                    : '',			
            'Quarter Day_upper':
                date.getDate() !== last_date.getDate()
					? date.getMonth() !== last_date.getMonth()
                    ? date_utils.format(date, 'D MMM YYYY', this.options.language)
					: date_utils.format(date, 'D', this.options.language)
                    : '',
            'Half Day_upper':
                date.getDate() !== last_date.getDate()
                    ? date.getMonth() !== last_date.getMonth()
                      ? date_utils.format(date, 'D MMM YYYY', this.options.language)
                      : date_utils.format(date, 'D', this.options.language)
                    : '',
            Day_upper:
                date.getMonth() !== last_date.getMonth()
                    ? date_utils.format(date, 'YYYY MMMM', this.options.language)
                    : '',
            Week_upper:
                date.getMonth() !== last_date.getMonth()
                    ? date_utils.format(date, 'YYYY MMMM', this.options.language)
                    : '',
            Month_upper:
                date.getFullYear() !== last_date.getFullYear()
                    ? date_utils.format(date, 'YYYY', this.options.language)
                    : '',
            Year_upper:
                date.getFullYear() !== last_date.getFullYear()
                    ? date_utils.format(date, 'YYYY', this.options.language)
                    : ''
        };

        const base_pos = {
            x: i * this.options.column_width + this.options.xoffset,
            lower_y: this.options.header_height,
            upper_y: this.options.header_height - 25
        };

        const x_pos = {
			'10Minutes_lower': 0,//this.options.column_width/2,
			'10Minutes_upper': this.options.column_width * 2/2,
			'30Minutes_lower': 0,//this.options.column_width/2,
			'30Minutes_upper': this.options.column_width * 2/2,
			'Hour_lower': 0,//this.options.column_width/2,
            'Hour_upper': this.options.column_width * 24/2,
            'Quarter Day_lower': 0, //this.options.column_width *1 / 2,
            'Quarter Day_upper': 0,
            'Half Day_lower': 0, //this.options.column_width * 1 / 2,
            'Half Day_upper': 0,
            Day_lower: 0,//this.options.column_width / 2,
            Day_upper: this.options.column_width * 30 / 2,
            Week_lower: 0,
            Week_upper: this.options.column_width * 4 / 2,
            Month_lower: this.options.column_width / 2,
            Month_upper: this.options.column_width * 12 / 2,
            Year_lower: this.options.column_width / 2,
            Year_upper: this.options.column_width * 30 / 2
        };
	//	console.log(date,this.options.view_mode,x_pos,base_pos,date_text)
        return {
			date: date,
            upper_text: date_text[`${this.options.view_mode}_upper`],
            lower_text: date_text[`${this.options.view_mode}_lower`],
            upper_x: base_pos.x + x_pos[`${this.options.view_mode}_upper`],
            upper_y: base_pos.upper_y,
            lower_x: base_pos.x + x_pos[`${this.options.view_mode}_lower`],
            lower_y: base_pos.lower_y
        };
    }

    make_bars() {
		let gantt_start = this.gantt_start;
		let gantt_end = this.gantt_end;
		let barlayer = this.layers.bar;
		this.bars = [];
		let that = this;
		this.tasks.forEach(function(task,index){
        //this.bars = this.tasks.map(task => {
			if(task._start < gantt_start)
				task._start = gantt_start;
			
			if(task._end > gantt_end)
				task._end = gantt_end;
			
            const bar = new Bar(that, task);
            barlayer.appendChild(bar.group);
            //return bar;
			that.bars.push(bar);
        });
		
	//	console.log(that.bars, that.tasks)
    }

    make_arrows() {
        this.arrows = [];
		let that = this;
		
		if(that.options.arrowtype == 'n' )  // without the link arrow line
		{
			return;
		}
		
		this.tasks.forEach(function(task, index){
        //for (let task of this.tasks) {
            let arrows = [];
			
			if( that.options.arrowtype == 's'){   // link to subtasks
				arrows = task.subtasks
					.map(task_id => {
						const subtask = that.get_task(task_id);
						if (!subtask) return;
						const arrow = new Arrow(
							that,
							that.bars[task._sequence], // from_task
							that.bars[subtask._sequence] // to_task
						  //  dependency.bar, // from_task
						  //  task.bar // to_task					
						);
						that.layers.arrow.appendChild(arrow.element);
						return arrow;
					})
					.filter(Boolean); // filter falsy values
				that.arrows = that.arrows.concat(arrows);				
			}
			else if( that.options.arrowtype == 'p' && task.parenttask !=''){   // link with parenttask
				if(task.parenttask =='')
					return;
				
				const arrow = new Arrow(
							that,
							that.bars[parenttask._sequence], // from_task
							that.bars[task._sequence] // to_task
						//	that.bars[subtask._sequence] // to_task
						  //  dependency.bar, // from_task
						  //  task.bar // to_task					
						);
				that.layers.arrow.appendChild(arrow.element);
				that.arrows.push(arrow); 
				
			}
			else {  // default, link with dependencies
				arrows = task.dependencies
					.map(task_id => {
						const dependency = that.get_task(task_id);
						if (!dependency) return;
						const arrow = new Arrow(
							that,
							that.bars[dependency._sequence], // from_task
							that.bars[task._sequence] // to_task
						  //  dependency.bar, // from_task
						  //  task.bar // to_task					
						);
						that.layers.arrow.appendChild(arrow.element);
						return arrow;
					})
					.filter(Boolean); // filter falsy values
				that.arrows = that.arrows.concat(arrows);
			}
			
        });
    }

    map_arrows_on_bars() {
        for (let bar of this.bars) {
            bar.arrows = this.arrows.filter(arrow => {
                return (
                    arrow.from_task.task.id === bar.task.id ||
                    arrow.to_task.task.id === bar.task.id
                );
            });
        }
    }

    set_width() {
		if(this.$svg == null) return;
		
        const cur_width = this.$svg.getBoundingClientRect().width;
        const actual_width = this.$svg
            .querySelector('.grid .grid-row')
            .getAttribute('width');
        if (cur_width < actual_width) {
            this.$svg.setAttribute('width', actual_width);
        }
		
		
		
    }
	get_scroll_position(){
		/*if(!this.scroll_position_date)
			return;
		*/
        const parent_element = this.$svg.parentElement;
        if (!parent_element) return;

		//this.scroll_position_date = this.get_date_of_point(parent_element.scrollLeft); //(parent_element.scrollLeft +  this.options.column_width) * this.options.column_width /this.options.step
		this.scroll_position_date = (parent_element.scrollLeft/this.options.column_width + 1) * this.options.step  //(parent_element.scrollLeft +  this.options.column_width) * this.options.column_width /this.options.step
	//	this.scrollLeft = parent_element.scrollLeft;
		
	//	console.log(parent_element.scrollLeft,this.options.column_width, this.options.step )
	}
    set_scroll_position() {
        const parent_element = this.$svg.parentElement;
        if (!parent_element) return;
		
		let scroll_start_hour =0;
		
		
		
		if(!this.scroll_position_date){
			const hours_before_first_task = date_utils.diff(
				this.get_oldest_starting_date(),
				this.gantt_start,
				'hour'
			);
			scroll_start_hour = hours_before_first_task;
			
		//	console.log(hours_before_first_task)
		/*	if(this.gantt_end > date_utils.today() && this.gantt_start < date_utils.today())
				scroll_start_hour = date_utils.diff(
					date_utils.today(),
					this.gantt_start,
					'hour'
				);
			else
				scroll_start_hour = hours_before_first_task;  */
		}
		else{
			scroll_start_hour = this.scroll_position_date;
		}
		
	//	console.log(this.scroll_position_date,scroll_start_hour)
		
			const scroll_pos =
				scroll_start_hour /
					this.options.step *
					this.options.column_width -
				this.options.column_width;

			parent_element.scrollLeft = scroll_pos;
	//	console.log(parent_element.scrollLeft,this.options.column_width, this.options.step )
    }
	scrool_to_time(datetime){
	    const parent_element = this.$svg.parentElement;
        if (!parent_element) return;
		
		let scroll_start_hour  = date_utils.diff(
					date_utils.parse(datetime),
					this.gantt_start,
					'hour'
				);
	//	console.log(scroll_start_hour);
		const scroll_pos =
				scroll_start_hour /
					this.options.step *
					this.options.column_width -
				this.options.column_width;

		parent_element.scrollLeft = scroll_pos;
	//	console.log(parent_element.scrollLeft,scroll_pos,this.options.column_width, this.options.step )
	}
	scroll_position(pos){
		const parent_element = this.$svg.parentElement;
        if (!parent_element) return;
		
		parent_element.scrollLeft = parent_element.scrollLeft + pos;
		
		return parent_element.scrollLeft;
	}

    bind_grid_click() {
        $.on(
            this.$svg,
            this.options.popup_trigger,
            '.grid-row, .grid-header',
            () => {
                this.unselect_all();
                this.hide_popup();
            }
        );
    }

    bind_bar_events() {
        let is_dragging = false;
        let x_on_start = 0;
        let y_on_start = 0;
        let is_resizing_left = false;
        let is_resizing_right = false;
        let parent_bar_id = null;
        let bars = []; // instanceof Bar
        this.bar_being_dragged = null;
		let dragging_bar = null;
		
		if(this.options.viewonly == 'yes')
			return;
		
        function action_in_progress() {
            return is_dragging || is_resizing_left || is_resizing_right;
        }

        $.on(this.$svg, 'mousedown', '.bar-wrapper, .handle', (e, element) => {
            const bar_wrapper = $.closest('.bar-wrapper', element);

            if (element.classList.contains('left')) {
                is_resizing_left = true;
            } else if (element.classList.contains('right')) {
                is_resizing_right = true;
            } else if (element.classList.contains('bar-wrapper')) {
                is_dragging = true;
            }

            bar_wrapper.classList.add('active');

            x_on_start = e.offsetX;
            y_on_start = e.offsetY;

            parent_bar_id = bar_wrapper.getAttribute('data-id');
            const ids = [
                parent_bar_id,
                ...this.get_all_dependent_tasks(parent_bar_id)
            ];
            bars = ids.map(id => this.get_bar(id));

            this.bar_being_dragged = parent_bar_id;
			this.dragging_bar =this.get_bar(this.bar_being_dragged)
			
            bars.forEach(bar => {
                const $bar = bar.$bar;
                $bar.ox = $bar.getX();
                $bar.oy = $bar.getY();
                $bar.owidth = $bar.getWidth();
                $bar.finaldx = 0;
            });
        });
		//let last_y = y_on_start;
		
        $.on(this.$svg, 'mousemove', e => {
            if (!action_in_progress()) return;
            const dx = e.offsetX - x_on_start;
            const dy = e.offsetY -  y_on_start;
			// only move the dragged bar from resource to resource
			
			if(this.dragging_bar != null)
				this.dragging_bar.update_bar_resource(this.get_destinate_resource(dy));
			
		//	console.log(bars)
            bars.forEach(bar => {
                const $bar = bar.$bar;
                $bar.finaldx = this.get_snap_position(dx);

                if (is_resizing_left && !bar.task.isstartviewonly) {
                    if (parent_bar_id === bar.task.id) {
                        bar.update_bar_position({							
                            x: $bar.ox + $bar.finaldx,
                            width: $bar.owidth - $bar.finaldx
                        });
                    } else {
                        bar.update_bar_position({
                            x: $bar.ox + $bar.finaldx
                        });
                    }
                } else if (is_resizing_right && !bar.task.isendviewonly) {
                    if (parent_bar_id === bar.task.id) {
                        bar.update_bar_position({
                            width: $bar.owidth + $bar.finaldx
                        });
                    }
                } else if (is_dragging) {
                    bar.update_bar_position({ x: $bar.ox + $bar.finaldx });				
                }
            });
        });

        document.addEventListener('mouseup', e => {
            if (is_dragging || is_resizing_left || is_resizing_right) {
                bars.forEach(bar => bar.group.classList.remove('active'));
            }

            is_dragging = false;
            is_resizing_left = false;
            is_resizing_right = false;
        });

        $.on(this.$svg, 'mouseup', e => {
			if(this.dragging_bar != null)
				this.dragging_bar.resource_change();
			
			if(this.dragging_bar != null && (is_dragging || is_resizing_left))
				this.dragging_bar.update_sub_tasks();
			
			
			this.dragging_bar = null
            this.bar_being_dragged = null;
			
            bars.forEach(bar => {
                const $bar = bar.$bar;
				
				//bar.resource_change();				
                
				if (!$bar.finaldx) return;
                bar.date_changed();		
								
                bar.set_action_completed();
            });
        });

        this.bind_bar_progress();
    }
	
	completedrawgantt(){
        this.trigger_event('gantt_completion_render', [
            this
        ]);		
	}

    bind_bar_progress() {
        let x_on_start = 0;
        let y_on_start = 0;
        let is_resizing = null;
        let bar = null;
        let $bar_progress = null;
        let $bar = null;

        $.on(this.$svg, 'mousedown', '.handle.progress', (e, handle) => {
            is_resizing = true;
            x_on_start = e.offsetX;
            y_on_start = e.offsetY;

            const $bar_wrapper = $.closest('.bar-wrapper', handle);
            const id = $bar_wrapper.getAttribute('data-id');
            bar = this.get_bar(id);

            $bar_progress = bar.$bar_progress;
            $bar = bar.$bar;

            $bar_progress.finaldx = 0;
            $bar_progress.owidth = $bar_progress.getWidth();
            $bar_progress.min_dx = -$bar_progress.getWidth();
            $bar_progress.max_dx = $bar.getWidth() - $bar_progress.getWidth();
        });

        $.on(this.$svg, 'mousemove', e => {
            if (!is_resizing) return;
            let dx = e.offsetX - x_on_start;
            let dy = e.offsetY - y_on_start;

            if (dx > $bar_progress.max_dx) {
                dx = $bar_progress.max_dx;
            }
            if (dx < $bar_progress.min_dx) {
                dx = $bar_progress.min_dx;
            }

            const $handle = bar.$handle_progress;
            $.attr($bar_progress, 'width', $bar_progress.owidth + dx);
            $.attr($handle, 'points', bar.get_progress_polygon_points());
            $bar_progress.finaldx = dx;
        });

        $.on(this.$svg, 'mouseup', () => {
            is_resizing = false;
            if (!($bar_progress && $bar_progress.finaldx)) return;
            bar.progress_changed();
            bar.set_action_completed();
        });
    }

	popup_change_data(data){		
		if(!data || data === null)
			return
		
		let task = this.get_task(data.taskid)
		let selectedbar = null
		selectedbar = this.get_bar(data.taskid)
		
		if(task._index != data.machine && data.machine != undefined && data.machine != ''){

			selectedbar.update_bar_resource(parseInt(data.machine)-task._index);			
		}else if(task.resource._index != data.workcenter && data.machine === '' && task.resource.parentid != ''){

			selectedbar.update_bar_resource(parseInt(data.workcenter)-task._index);			
		}
		
        let start = date_utils.parse(new Date(data.start));
        let end = date_utils.parse(new Date(data.end));	

		const parent_bar_id = task.id
		const ids = [
                parent_bar_id,
                ...this.get_all_dependent_tasks(parent_bar_id)
            ];
        const bars = ids.map(id => this.get_bar(id));
		
        bars.forEach(bar => {
                const $bar = bar.$bar;
                $bar.ox = $bar.getX();
                $bar.oy = $bar.getY();
                $bar.owidth = $bar.getWidth();
                $bar.finaldx = 0;
        });
	//	console.log(bars)
		
		let deltastart = date_utils.diff(start, task._start, 'minutes')
		let deltaend = date_utils.diff(end, task._end, 'minutes')
	//	console.log(start,end,deltastart,deltaend)
	//	console.log(start,end,deltastart,deltaend,date_utils.add(selectedbar.task._start, deltastart, 'minute'),date_utils.add(selectedbar.task._end, deltaend, 'minute'))
		bars.forEach(bar => {
				const $bar = bar.$bar;
				//$bar.finaldx = 1;
				bar.update_bar_position({
                        start:date_utils.add(bar.task._start, deltastart, 'minute'),
						end:date_utils.add(bar.task._end, deltaend, 'minute')						
                    });
					
				/*if(bar.task.resource.type == 'line'){
					bar.update_sub_tasks();
				}*/	
				bar.date_changed();
				bar.set_action_completed();		
		})
	
	
    /*    bars.forEach(bar => {
            const $bar = bar.$bar;				
				//bar.resource_change();				
                
			//if (!$bar.finaldx) return;
                bar.date_changed();		
								
            bar.set_action_completed();
        }); 
		*/
				
		selectedbar.resource_change();
		this.hide_popup();
	}
	
	remove_task_badge(taskid,employeeid){
	//	console.log(taskid,employeeid)
		let selectedbar = this.get_bar(taskid);
	//	console.log(selectedbar)
		selectedbar.remove_badge(employeeid);
	}
	
    get_all_dependent_tasks(task_id) {
        let out = [];
        let to_process = [task_id];
        while (to_process.length) {
            const deps = to_process.reduce((acc, curr) => {
                acc = acc.concat(this.dependency_map[curr]);
                return acc;
            }, []);

            out = out.concat(deps);
            to_process = deps.filter(d => !to_process.includes(d));
        }

        return out.filter(Boolean);
    }

	get_destinate_resource(dy){
		const row_height = this.options.bar_height + this.options.padding;
		let ody = dy,
			rem,
			position;
			
		rem = dy % row_height
		position = (dy - rem) / row_height + (
			rem < row_height /2 ? 0 : 1
		)
		return position
		
	}
	
	update_task_issue(taskid, issuetype){
		let bar = this.get_bar(taskid);
		bar.task.issuetype = issuetype;
		
		bar.draw_issue_icon();
		
	}
	
	draw_new_badges(taskid, badgetext, badgeid, description){
		
		let bar = this.get_bar(taskid);
		//console.log(bar)
		if(bar == null)
			return;

		let r = (bar.height - 4)/2
		let y = bar.$bar.getY() + bar.height/2; 

		
		let badges = [];
		// badges: employee
        if (typeof bar.task.badges === 'string' || !bar.task.badges) {
            if (bar.task.badges) {
                    badges = bar.task.badges
                        .split(',')
                        .map(d => d.trim())
                        .filter(d => d);
                }
        }

		let x = bar.$bar.getX() + 10 + (badges.length+1) * (r + 2);		

		let bar_badge = createSVG('circle', {
							cx: x,
							cy: y,
							r: r,
							adid: badgeid,
							title: description,
							class: 'bar-badge',
							append_to: bar.bar_group
						});
						
		let bar_badge_label = createSVG('text', {
							x: x,
							y: y,
							innerHTML: badgetext,
							class: 'bar-badge-label',
							append_to: bar.bar_group
						});	
		
		if (typeof bar.task.badges === 'string' || !bar.task.badges)
			bar.task.badges = bar.task.badges + ',' + badgeid +'_' + badgetext+'_'+description;
		else	
			bar.task.badges = ',' + badgeid +'_' + badgetext+'_'+description;

	}
	
	get_position_workingtime(x, direction){
		let date = this.get_date_of_point(x);
		
		if(date.getDate() === 6){
			date = date_utils.add(date.setHours(0,0,0,0),direction==1? 2: -1,'hour');
		}else if(date.getDate() === 0){
			date = date_utils.add(date.setHours(0,0,0,0),direction==1? 1: -2,'hour');
		}
		
		return this.options.xoffset + date_utils.diff(date,this.gantt_start,'hour') / this.options.step * this.options.column_width
	}
	
	get_date_of_point(x){
		return date_utils.add(this.gantt_start,60*(this.options.step*(x - this.options.xoffset)/this.options.column_width),'minute')
	}
	get_point_bydate(date){
		return this.options.xoffset + date_utils.diff(date,this.gantt_start,'minute') /60 / this.options.step * this.options.column_width 
	}
	
    get_snap_position(dx) {
        let odx = dx,
            rem,
            position;
		
		return dx;
		
        if (this.view_is('Week')) {
            rem = dx % (this.options.column_width / 7);
            position =
                odx -
                rem +
                (rem < this.options.column_width / 14
                    ? 0
                    : this.options.column_width / 7);
        } else if (this.view_is('Month')) {
            rem = dx % (this.options.column_width / 30);
            position =
                odx -
                rem +
                (rem < this.options.column_width / 60
                    ? 0
                    : this.options.column_width / 30);
        } else {
            rem = dx % this.options.column_width;
            position =
                odx -
                rem +
                (rem < this.options.column_width / 2
                    ? 0
                    : this.options.column_width);
        }
		
		
		
        return position;
    }
	
	check_overlap_bar(){
		let delta = 0;
		
		console.log(this.resources)
		
		 for (let resource of this.resources){ 
			this.check_overlap_for_resource(resource.id)
			
		 }
	};
	
	check_overlap_for_bar(sbar){
		let delta = 0;
		
		var overlape = false;
		for (let tbar of this.bars){ 

			if(sbar.task.resource.id == tbar.task.resource.id){
				//console.log((parseInt(sbar.x)+ parseInt(sbar.width)), parseInt(tbar.x) + delta,sbar.y, tbar.y, (parseInt(sbar.x) +delta), (parseInt(tbar.x) + parseInt(tbar.width)))
				if((sbar.$bar.getEndX()) > parseInt(tbar.$bar.getX()) + delta  && (parseInt(sbar.$bar.getX()) +delta) < parseInt(tbar.$bar.getEndX()) && tbar != sbar){
						overlape = true;
						
					//	console.log('overlap',sbar, tbar,this.options)
						
						if(sbar.height == this.options.bar_height){
							sbar.y = sbar.compute_y();
							sbar.height = this.options.bar_height/2 -1;
							sbar.update_bar_height();
						}else if(sbar.$bar.getY() == tbar.$bar.getY() && sbar.height == tbar.height){
							if(sbar.$bar.getY() == sbar.compute_y()){
								sbar.y = sbar.compute_y()+ this.options.bar_height/2 + 2;								
							}
							else 
							   sbar.y = sbar.compute_y();
						   
							sbar.update_bar_height();
						}
											
						if(tbar.height == this.options.bar_height){
						//	console.log(tbar)
							tbar.height = this.options.bar_height/2;	

							if(sbar.y == tbar.$bar.getY())
								tbar.y = tbar.$bar.getY()+ this.options.bar_height/2 + 2;
														
							tbar.update_bar_height();
						}
					
				}
			}			
				
		}
			
		if(overlape == false && parseInt(sbar.height) < this.options.bar_height){
				sbar.height = this.options.bar_height;
				sbar.y = sbar.compute_y();
				sbar.update_bar_height();
		}		
		
	};
	
	check_overlap_for_resource(resourceid){
		for (let bar of this.bars){		
			if(bar.task.resource.id == resourceid)
				this.check_overlap_for_bar(bar)
			
		 }
		
		
	}
	
    unselect_all() {
        [...this.$svg.querySelectorAll('.bar-wrapper')].forEach(el => {
            el.classList.remove('active');
        });
    }

    view_is(modes) {
        if (typeof modes === 'string') {
            return this.options.view_mode === modes;
        }

        if (Array.isArray(modes)) {
            return modes.some(mode => this.options.view_mode === mode);
        }

        return false;
    }

    get_task(id) {
        return this.tasks.find(task => {
            return task.id === id;
        });
    }

    get_bar(id) {
        return this.bars.find(bar => {
            return bar.task.id === id;
        });
    }
	
    show_popup(options) {
        if (!this.popup) {
            this.popup = new Popup(
                this.popup_wrapper,
                this.options.custom_popup_html
            );
        }
        this.popup.show(options);
		
		this.popup_wrapper.style.left = (options.e.offsetX + 20) +"px";
		this.popup_wrapper.style.top = (options.e.offsetY -10)+"px";
		//$(".popup-wrapper").css("left", options.e.offsetX)
		//$(".popup-wrapper").css("top", options.e.offsetY)
		//$('div.popup-wrapper').show();
    }

	
    hide_popup() {
        this.popup && this.popup.hide();
    }

    trigger_event(event, args) {
        if (this.options['on_' + event]) {
            this.options['on_' + event].apply(null, args);
        }
    }

    /**
     * Gets the oldest starting date from the list of tasks
     *
     * @returns Date
     * @memberof Gantt
     */
    get_oldest_starting_date() {
		if(this.tasks.length > 0)
			return this.tasks
				.map(task => task._start)
				.reduce(
					(prev_date, cur_date) =>
						cur_date <= prev_date ? cur_date : prev_date
				);
		else
			return this.gantt_start;
    }

    /**
     * Clear all elements from the parent svg element
     *
     * @memberof Gantt
     */
    clear() {
        this.$svg.innerHTML = '';
    }
}

function generate_id(task) {
    return (
        task.name +
        '_' +
        Math.random()
            .toString(36)
            .slice(2, 12)
    );
}

return Gantt;

}());
