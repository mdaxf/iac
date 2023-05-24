function read_from_url(url)
{
	if (waiting) return;
	waiting=true;



	if (url.length<1)
	{
		//alert('Please enter a valid url');
		return after_error();
	}
	url_is_local = true;
	if (url_is_local) return download_from_local(url, url, true);
				
	var xhr = new XMLHttpRequest();
				
	xhr.onreadystatechange =
	function(e)
	{
		if (xhr.readyState == 4)
		{
			var pos;
			var s=xhr.responseText.trim();
			//console.log(s);
			if (s.substr(0,2)=='OK')
			{
				pos=s.indexOf('~');
				s=s.substr(pos+1);
				pos=s.indexOf('~');
				var temp_filesize=s.substr(0, pos);
							
				s=s.substr(pos+1);
				pos=s.indexOf('~');
				var temp_orig_filename=s.substr(0, pos);
							
				s=s.substr(pos+1);
				var temp_filename=((s.length<1)?"unknown":s);
						
				download_from_local(temp_filename, temp_orig_filename);
			}
			else
			{
				if (s.substr(0,2)=='E1')
					alert('Invalid link');
				else if (s.substr(0,2)=='E2')
					alert('File is too big - maximum allowed size is 30mb');
				else if (s.substr(0,2)=='E4')
					alert('Unknown file type');
				else
					alert('Error reading from url');
					
				return after_error();
			}
		}
	}
	x.open("GET", url, true);
	xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
	x.send(null);
	/*xhr.open("POST", "/general/read_from_url.php", true);
	xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
	xhr.send("url="+url);  */
}


function download_from_local(filename, orig_filename, is_local)
{
	is_local=is_local||false;
	
	var pbar=$id('file_pbar');
	var xhr = new XMLHttpRequest();

	xhr.onprogress =
	function(e)
	{
		if (cancel_download)
		{
			del_temp_file(filename);
			xhr.abort();
			return after_error();
		}
		else
			if (pbar!=null) pbar.value=e.loaded / e.total*100;
	}
				
	xhr.onreadystatechange =
	function(e)
	{
		if (!cancel_download)
			if (xhr.readyState == 4)
			{
				del_temp_file(filename);
				switch_view('proc');
				setTimeout(function(){after_file_load(orig_filename, xhr.response)}, 500);
			}
	}
	
	pbar.value=0;
	switch_view('pbar');	
	
	xhr.open((is_local?"GET":"POST"), filename, true);
	xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
	xhr.responseType = "arraybuffer";
	xhr.send(null);
}

function del_temp_file(f)
{
	var xhr = new XMLHttpRequest();
	/*			
	xhr.open("POST", "/general/del_temp.php", true);
	xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
	xhr.send("filename="+f);  */
}
