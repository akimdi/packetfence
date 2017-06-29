package pf::detect::parser::streamscan;

=head1 NAME

pf::detect::parser::streamscan

=cut

=head1 DESCRIPTION

pf::detect::parser::streamscan

Class to parse syslog from StreamScan

=cut

use strict;
use warnings;
use Moo;
extends qw(pf::detect::parser);

sub parse {
    my ($self,$line) = @_;
    my $data;
    if ( $line
        =~ /CDS\[(?<cds_id>\d+)\].*?type=(?<type>[^ ]*).*?threat="(?<threat>.*?)" direction=(?<direction>[^ ]+) sourceip=(?<sourceip>\d+(\.\d+){3}) sourceport=(?<sourceport>\d+) destip=(?<ip>\d+(\.\d+){3}) destport=(?<destport>\d+) app=(?<app>[^ ]*) timestamp=(?<timestamp>[^ ]*) sid=(?<sid>\d+)/
        )
    {
    
        $data = {
            date  => $+{timestamp},
            sid   => $+{sid},
            descr => $+{threat},
            srcip => $+{sourceip},
            dstip => $+{destip},
        };
        return { srcip => $data->{srcip}, date => $data->{date}, dstip => $data->{dstip}, events => { detect => $data->{sid}, suricata_event => $data->{descr} } };
    }
    return undef;
}
 
=head1 AUTHOR

Inverse inc. <info@inverse.ca>

=head1 COPYRIGHT

Copyright (C) 2005-2017 Inverse inc.

=head1 LICENSE

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301,
USA.

=cut

1;
