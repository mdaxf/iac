function upload_file(f)
{
	if (waiting) return;
				
	if (!supported_file_type(f.name))
	{
		alert('File type is not supported');
		return false;
	}
				
				
	if (f.size>max_file_size)
	{
		alert('File is too big - maximum allowed size is 20mb');
		return false;
	}					
				
	read_file(f);
}



function read_file(f)
{
	waiting=true;
	
	var pbar=$id('file_pbar');
	var reader = new FileReader();
				
	reader.onerror = function(e)
	{
		var error_str="";
		switch(e.target.error.code)
		{
			case e.target.error.NOT_FOUND_ERR:
				error_str="File not found";
			break;

			case e.target.error.NOT_READABLE_ERR:
				error_str="Can't read file - too large?";
			break;

			case e.target.error.ABORT_ERR:
				error_str="Read operation aborted";
			break; 
						
			case e.target.error.SECURITY_ERR:
				error_str="File is locked";
			break;

			case e.target.error.ENCODING_ERR:
				error_str="File too large";

			break;

			default:
				error_str="Error reading file";
		}
		alert(error_str);
		switch_view('drag');
		return after_error();
	}       
					
	reader.onload = function(e)
	{
		switch_view('proc');
		setTimeout(function(){after_file_load(f.name, e.target.result)}, 500);
	};
				
	reader.onprogress = function(e)
	{
		if (cancel_download)
		{
			reader.abort();
			return after_error();
		}
		else	
			pbar.value=e.loaded / e.total*100;
	};	
				
	pbar.value=0;
				
	switch_view('pbar');
				
	reader.readAsArrayBuffer(f);
}